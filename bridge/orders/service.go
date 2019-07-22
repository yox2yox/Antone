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

// type OrderResult struct {
// 	Db         int32
// 	IsRejected bool
// 	IsError    bool
// }

// type Order struct {
// 	HolderId       string
// 	CreatedTime    time.Time
// 	WaitingWorkers []string
// 	OrderResults   map[string]OrderResult
// }

// func NewOrder(holderId string) *Order {
// 	return &Order{
// 		HolderId:       holderId,
// 		CreatedTime:    time.Now(),
// 		WaitingWorkers: []string{},
// 		OrderResults:   map[string]OrderResult{},
// 	}
// }

// func (o *Order) AddWorkers(workersId []string) {
// 	o.WaitingWorkers = append(o.WaitingWorkers, workersId...)
// }

type Service struct {
	sync.RWMutex
	//WaitList                    []*Order //TODO:Remove?
	ValidationRequests          []*ValidationRequest
	Accounting                  *accounting.Service
	Running                     bool
	WithoutConnectRemoteForTest bool
	stopChan                    chan struct{}
}

func NewService(accounting *accounting.Service, withoutConnectRemoteForTest bool) *Service {
	return &Service{
		//WaitList:                    []*Order{},
		ValidationRequests:          []*ValidationRequest{},
		Accounting:                  accounting,
		Running:                     false,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
		stopChan:                    make(chan struct{}),
	}
}

//TODO:Remove Order WaitList
// func (s *Service) GetOrders() []*Order {
// 	return s.WaitList
// }

func (s *Service) getValidatableCodeRemote(holder *accounting.Worker, userId string, add int32) (*pb.ValidatableCode, error) {
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
func (s *Service) GetValidatableCode(userId string, add int32) (*pb.ValidatableCode, string, error) {
	holder, err := s.Accounting.SelectDataPoolHolder(userId)
	if err != nil {
		return nil, "", err
	}
	var vcode *pb.ValidatableCode
	if !s.WithoutConnectRemoteForTest {
		vcode, err = s.getValidatableCodeRemote(holder, userId, add)
		if err != nil {
			return nil, holder.Id, err
		}
	} else {
		vcode = &pb.ValidatableCode{
			Data: 0,
			Add:  add,
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

func (s *Service) AddValidationRequest(needNum int, holderId string, vcode *pb.ValidatableCode) {
	vreq := &ValidationRequest{Neednum: needNum, HolderId: holderId, ValidatableCode: vcode}
	s.Lock()
	s.ValidationRequests = append(s.ValidationRequests, vreq)
	s.Unlock()
}

//結果リストから、必要な追加ワーカ数および最も多数だった結果を返す
func (s *Service) calcNeedAdditionalWorkerAndResult(picknum int, results map[string]*ValidationResult) (int, *ValidationResult) {
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
	conclusion := &ValidationResult{}
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

//ValidatableCodeを検証
func (s *Service) ValidateCode(ctx context.Context, picknum int, holderId string, vCode *pb.ValidatableCode) error {

	waitlist := []string{}
	results := map[string]*ValidationResult{}

	errChan := make(chan error)                //エラー送信用チャネル
	resultChan := make(chan *ValidationResult) //バリデーション結果送信チャネル
	done := make(chan struct{})                //完了チャネル

	go func() {
		nextpick := picknum

		for {
			if nextpick > 0 {
				//ワーカー取得
				pickedWorkers, err := s.Accounting.SelectValidationWorkers(nextpick)

				if err == nil {
					nextpick = 0
					pickedIDs := []string{}
					for _, worker := range pickedWorkers {
						pickedIDs = append(pickedIDs, worker.Id)
					}
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
				results[res.WorkerId] = res
				needworker, _ := s.calcNeedAdditionalWorkerAndResult(picknum, results)
				if needworker > 0 {
					nextpick = needworker - len(waitlist)
					if nextpick < 0 {
						nextpick = 0
					}
				} else {
					close(done)
					return
				}
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case <-done:
			for _, res := range results {
				s.Accounting.UpdateReputation(res.WorkerId, !res.IsRejected, res.IsError)
			}
			return nil
		}
	}
}

func (s *Service) Run() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			if len(s.ValidationRequests) > 0 {
				fmt.Println("get item")
				vreq := s.dequeueValidationRequest()
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					s.ValidateCode(ctx, vreq.Neednum, vreq.HolderId, vreq.ValidatableCode)
					//TODO:結果を記録
				}()
			}
		}
	}
}

func (s *Service) Stop() {
	close(s.stopChan)
}
