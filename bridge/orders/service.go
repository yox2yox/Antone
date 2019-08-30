package orders

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"yox2yox/antone/bridge/accounting"
	pb "yox2yox/antone/bridge/pb"
	workerpb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

var (
	ErrGottenDataIsNil = errors.New("You got data but it's nil")
)

type ValidationRequest struct {
	UserId          string
	Neednum         int
	HolderId        string
	ValidatableCode *pb.ValidatableCode
}

type ValidationResult struct {
	WorkerId   string
	Data       int32
	IsRejected bool
	IsError    bool
}

type Service struct {
	sync.RWMutex
	ValidationRequests          []*ValidationRequest
	WaitingRequestsId           []string
	Accounting                  *accounting.Service
	Running                     bool
	WithoutConnectRemoteForTest bool
	stopChan                    chan struct{}
}

func NewService(accounting *accounting.Service, withoutConnectRemoteForTest bool) *Service {
	return &Service{
		ValidationRequests:          []*ValidationRequest{},
		Accounting:                  accounting,
		Running:                     false,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
		stopChan:                    make(chan struct{}),
	}
}

func (s *Service) IsTheDatapoolAvailable(userId string) bool {
	s.RLock()
	available := true
	for _, id := range s.WaitingRequestsId {
		if id == userId {
			available = false
		}
	}
	s.RUnlock()
	return available
}

func (s *Service) getValidatableCodeRemote(holder accounting.Worker, userId string, add int32) (*pb.ValidatableCode, error) {
	conn, err := grpc.Dial(holder.Addr, grpc.WithInsecure())
	client := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &workerpb.ValidatableCodeRequest{Userid: userId, Add: add}
	vcode, err := client.GetValidatableCode(ctx, req)
	if err != nil {
		return nil, err
	}
	if vcode == nil {
		return nil, ErrGottenDataIsNil
	}
	vcodeout := &pb.ValidatableCode{Data: vcode.Data, Add: vcode.Add}
	return vcodeout, nil

}

//ValidatableCodeを取得する
func (s *Service) GetValidatableCode(ctx context.Context, userId string, add int32) (*pb.ValidatableCode, string, error) {
	//データプールが利用可能になるまで無限ループ
	for s.IsTheDatapoolAvailable(userId) == false {
		select {
		case <-ctx.Done():
			return nil, "", nil
		default:
		}
	}
	holder, err := s.Accounting.SelectDataPoolHolder(userId, []string{})
	if err != nil {
		return nil, "", err
	}
	var vcode *pb.ValidatableCode
	if s.WithoutConnectRemoteForTest {
		vcode = &pb.ValidatableCode{
			Data: 0,
			Add:  add,
		}
	} else {
		vcode, err = s.getValidatableCodeRemote(holder, userId, add)
		if err != nil {
			return nil, holder.Id, err
		}
	}
	return vcode, holder.Id, nil
}

func (s *Service) validateCodeRemote(worker *accounting.Worker, vCode *pb.ValidatableCode) *ValidationResult {
	conn, err := grpc.Dial(worker.Addr, grpc.WithInsecure())
	workerClient := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	vCodeWorker := &workerpb.ValidatableCode{Data: vCode.Data, Add: vCode.Add}
	validationResult, err := workerClient.OrderValidation(ctx, vCodeWorker)
	if err != nil {
		return &ValidationResult{WorkerId: worker.Id, Data: 0, IsRejected: false, IsError: true}
	} else {
		return &ValidationResult{WorkerId: worker.Id, Data: validationResult.Pool, IsRejected: false, IsError: false}
	}
}

func (s *Service) dequeueValidationRequest() *ValidationRequest {
	s.Lock()
	tmp := s.ValidationRequests[0]
	s.ValidationRequests = s.ValidationRequests[1:]
	s.Unlock()
	return tmp
}

func (s *Service) AddValidationRequest(userId string, needNum int, holderId string, vcode *pb.ValidatableCode) {
	fmt.Printf("DEBUG %s [] A ValidationRequest is added to orders.service\n", time.Now().String())
	vreq := &ValidationRequest{UserId: userId, Neednum: needNum, HolderId: holderId, ValidatableCode: vcode}
	s.Lock()
	s.ValidationRequests = append(s.ValidationRequests, vreq)
	s.Unlock()
	fmt.Printf("DEBUG %s [] Waiting ValidationRequests count is %d\n", time.Now().String(), len(s.ValidationRequests))
}

//waitlistに追加
func (s *Service) addWaitingList(userId string) {
	s.Lock()
	s.WaitingRequestsId = append(s.WaitingRequestsId, userId)
	s.Unlock()
}

//waitlistから削除
func (s *Service) removeWaitingList(userId string) {
	newlist := []string{}
	s.RLock()
	for _, waitId := range s.WaitingRequestsId {
		if waitId != userId {
			newlist = append(newlist, waitId)
		}
	}
	s.RUnlock()

	s.Lock()
	s.WaitingRequestsId = newlist
	s.Unlock()
}

//結果リストから、必要な追加ワーカ数および最も多数だった結果を返す
func (s *Service) calcNeedAdditionalWorkerAndResult(picknum int, results []ValidationResult) (int, ValidationResult) {
	//TODO: 確率を計算する
	resultmap := map[int32]int{}
	rejects := 0
	for _, res := range results {
		if res.IsRejected {
			rejects += 1
		} else if !res.IsError {
			_, exist := resultmap[res.Data]
			if exist {
				resultmap[res.Data] += 1
			} else {
				resultmap[res.Data] = 1
			}
		}
	}

	//TODO: 同率の場合の処理
	maxcount := -1
	resdata := int32(0)
	for data, count := range resultmap {
		if count > maxcount {
			maxcount = count
			resdata = data
		}
	}

	var need int
	conclusion := ValidationResult{}
	if maxcount > rejects {
		conclusion.IsRejected = false
		conclusion.IsError = false
		conclusion.Data = resdata
		need = picknum - maxcount
	} else {
		conclusion.IsRejected = true
		conclusion.IsError = false
		conclusion.Data = 0
		need = picknum - rejects
	}

	if need < 0 {
		need = 0
	}

	half := s.Accounting.GetWorkersCount()/2 + 1
	needhalf := half - maxcount
	if needhalf < 0 {
		needhalf = 0
	}
	if need > needhalf {
		need = needhalf
	}
	return need, conclusion

}

//バリデーション結果に基づいて評価値を登録する
func (s *Service) commitReputation(holderId string, results []ValidationResult, conclusion *ValidationResult) {
	if conclusion != nil {
		if conclusion.IsRejected == true {
			s.Accounting.UpdateReputation(holderId, false)
		} else {
			s.Accounting.UpdateReputation(holderId, true)
		}
	}
	for _, res := range results {
		confirmed := false
		if res.Data == conclusion.Data && res.IsError == false {
			confirmed = true
		}
		if res.IsRejected == true && res.IsError == false && conclusion.IsRejected == true {
			confirmed = true
		}
		s.Accounting.UpdateReputation(res.WorkerId, confirmed)
	}
}

//ValidatableCodeを検証
func (s *Service) ValidateCode(ctx context.Context, picknum int, holderId string, vCode *pb.ValidatableCode) ([]ValidationResult, *ValidationResult, error) {

	anavailable := []string{}
	waitlist := []string{}
	results := []ValidationResult{}
	var conclusion *ValidationResult

	errChan := make(chan error)                //エラー送信用チャネル
	resultChan := make(chan *ValidationResult) //バリデーション結果送信チャネル

	nextpick := picknum

	for {
		if nextpick > 0 {
			//ワーカー取得
			pickedWorkers, err := s.Accounting.SelectValidationWorkers(nextpick, anavailable)

			if err == nil {
				nextpick = 0
				pickedIDs := []string{}
				for _, worker := range pickedWorkers {
					pickedIDs = append(pickedIDs, worker.Id)
				}
				anavailable = append(anavailable, pickedIDs...)
				waitlist = append(waitlist, pickedIDs...)

				//各ワーカにValidationリクエスト送信
				for _, worker := range pickedWorkers {
					if s.WithoutConnectRemoteForTest == false {
						go func(target *accounting.Worker) {
							validationResult := s.validateCodeRemote(target, vCode)
							resultChan <- validationResult
						}(worker)
					} else {
						go func(target *accounting.Worker) {
							tmp := []string{}
							for _, id := range waitlist {
								if id != target.Id {
									tmp = append(tmp, id)
								}
							}
							waitlist = tmp
							resultChan <- &ValidationResult{WorkerId: target.Id, Data: 10, IsRejected: false, IsError: false}
						}(worker)
					}
				}
			} else {
				errChan <- err
			}
		}

		select {
		case res := <-resultChan:
			results = append(results, *res)
			var needworker int
			needworker, concval := s.calcNeedAdditionalWorkerAndResult(picknum, results)
			conclusion = &concval
			if needworker > 0 {
				conclusion = nil
				nextpick = needworker - len(waitlist)
				if nextpick < 0 {
					nextpick = 0
				}
			} else {
				return results, conclusion, nil
			}
		case <-ctx.Done():
			return results, nil, ctx.Err()
		case err := <-errChan:
			return results, nil, err
		default:
		}
	}
}

func (s *Service) Run() {
	for {
		select {
		case <-s.stopChan:
			fmt.Printf("INFO %s [] OrdersServer has been stoped\n", time.Now())
			return
		default:
			if len(s.ValidationRequests) > 0 {
				fmt.Printf("DEBUG %s [] Got order to validate\n", time.Now())

				//利用可能なリクエストが出てくるまでループ
				waitingvreq := []*ValidationRequest{}
				available := false
				var vreq *ValidationRequest = nil
				for available == false {
					vreq = s.dequeueValidationRequest()
					available = s.IsTheDatapoolAvailable(vreq.UserId)
					if available == false {
						waitingvreq = append(waitingvreq, vreq)
					}
				}

				if vreq != nil && available == true {
					//WaitingListに追加
					s.addWaitingList(vreq.UserId)
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
						defer cancel()
						results, conclusion, _ := s.ValidateCode(ctx, vreq.Neednum, vreq.HolderId, vreq.ValidatableCode)
						//結果から評価値を登録
						if results != nil {
							fmt.Printf("INFO %s [] END - Validations by remote workers", time.Now())
							s.commitReputation(vreq.HolderId, results, conclusion)
						}
						//結果を送信
						if conclusion != nil && conclusion.IsRejected == false {
							s.Accounting.UpdateDatapoolRemote(vreq.UserId, conclusion.Data)
						}
						//WaingListから削除
						s.removeWaitingList(vreq.UserId)
					}()
				} else {
					//取り出したリクエストを待ち行列に戻す
					s.Lock()
					s.ValidationRequests = append(waitingvreq, s.ValidationRequests...)
					s.Unlock()
				}
			}
		}
	}
}

func (s *Service) Stop() {
	close(s.stopChan)
}
