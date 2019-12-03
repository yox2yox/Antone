package orders

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"yox2yox/antone/bridge/accounting"
	"yox2yox/antone/bridge/datapool"
	pb "yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/pkg/stolerance"
	workerpb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

var (
	ErrGottenDataIsNil        = errors.New("You got data but it's nil")
	ErrFailedToCalcNeedWorker = errors.New("Failed to calclate need worker count")
)

type ValidationRequest struct {
	DatapoolId      string
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
	Datapool                    *datapool.Service
	Running                     bool
	WithoutConnectRemoteForTest bool
	CalcER                      bool
	ErrCount                    int
	FailedCount                 int
	SuccessCount                int
	RejectedCount               int
	WorksCount                  int
	OnBridgeCount               int
	NodesAvg                    float64
	NodesSum                    int
	stopChan                    chan struct{}
}

func NewService(accounting *accounting.Service, datapool *datapool.Service, withoutConnectRemoteForTest bool, calcER bool) *Service {
	return &Service{
		ValidationRequests:          []*ValidationRequest{},
		Accounting:                  accounting,
		Datapool:                    datapool,
		Running:                     false,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
		CalcER:                      calcER,
		ErrCount:                    0,
		SuccessCount:                0,
		FailedCount:                 0,
		RejectedCount:               0,
		OnBridgeCount:               0,
		stopChan:                    make(chan struct{}),
	}
}

func (s *Service) IsTheDatapoolAvailable(datapoolId string) bool {
	s.RLock()
	available := true
	for _, id := range s.WaitingRequestsId {
		if id == datapoolId {
			available = false
		}
	}
	s.RUnlock()
	return available
}

func (s *Service) getWaitingValidationRequestsCount() int {
	return len(s.ValidationRequests)
}

func (s *Service) getValidatableCodeRemote(holder accounting.Worker, datapoolId string, add int32) (*pb.ValidatableCode, error) {
	conn, err := grpc.Dial(holder.Addr, grpc.WithInsecure())
	client := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer conn.Close()
	defer cancel()
	req := &workerpb.ValidatableCodeRequest{Datapoolid: datapoolId, Add: add}
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
func (s *Service) GetValidatableCode(ctx context.Context, datapoolId string, add int32) (*pb.ValidatableCode, string, error) {
	//データプールが利用可能になるまで無限ループ
	for s.IsTheDatapoolAvailable(datapoolId) == false {
		select {
		case <-ctx.Done():
			return nil, "", nil
		default:
		}
	}
	holder, err := s.Datapool.SelectDataPoolHolder(datapoolId, []string{})
	if err != nil {
		return nil, "", err
	}
	var vcode *pb.ValidatableCode = nil
	if s.WithoutConnectRemoteForTest {
		vcode = &pb.ValidatableCode{
			Data: 0,
			Add:  add,
		}
	} else {
		vcode, err = s.getValidatableCodeRemote(*holder, datapoolId, add)
		if err != nil {
			return nil, holder.Id, err
		}
	}
	if vcode == nil {
		return nil, holder.Id, ErrGottenDataIsNil
	} else {
		return vcode, holder.Id, nil
	}
}

func (s *Service) validateCodeRemote(worker *accounting.Worker, vCode *pb.ValidatableCode, badreputations []int32) *ValidationResult {
	conn, err := grpc.Dial(worker.Addr, grpc.WithInsecure())
	if err != nil {
		log2.Err.Printf("failed to connect to remote %#v", err)
	}
	defer conn.Close()
	workerClient := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	vCodeWorker := &workerpb.ValidatableCode{Data: vCode.Data, Add: vCode.Add, Reputation: int32(worker.Reputation), Badreputations: badreputations, Threshould: float32(s.Accounting.CredibilityThreshould), Resetrate: float32(s.Accounting.ReputationResetRate)}
	validationResult, err := workerClient.OrderValidation(ctx, vCodeWorker)
	if err != nil {
		log2.Err.Printf("failed to validation on remote %#v", err)
		return &ValidationResult{WorkerId: worker.Id, Data: 0, IsRejected: false, IsError: true}
	} else {
		return &ValidationResult{WorkerId: worker.Id, Data: validationResult.Pool, IsRejected: false, IsError: false}
	}
}

func (s *Service) dequeueValidationRequest() *ValidationRequest {
	s.RLock()
	count := len(s.ValidationRequests)
	s.RUnlock()
	if count <= 0 {
		return nil
	}
	s.Lock()
	tmp := s.ValidationRequests[0]
	if len(s.ValidationRequests) > 1 {
		s.ValidationRequests = s.ValidationRequests[1:]
	} else {
		s.ValidationRequests = []*ValidationRequest{}
	}
	s.Unlock()
	return tmp
}

func (s *Service) AddValidationRequest(datapoolId string, needNum int, holderId string, vcode *pb.ValidatableCode) {
	fmt.Printf("DEBUG %s [] A ValidationRequest is added to orders.service\n", time.Now().String())
	vreq := &ValidationRequest{DatapoolId: datapoolId, Neednum: needNum, HolderId: holderId, ValidatableCode: vcode}
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
	s.Lock()
	for _, waitId := range s.WaitingRequestsId {
		if waitId != userId {
			newlist = append(newlist, waitId)
		}
	}
	s.WaitingRequestsId = newlist
	s.Unlock()
}

//結果リストから、必要な追加ワーカ数および最も多数だった結果を返す
func (s *Service) calcNeedAdditionalWorkerAndResult(picknum int, results []ValidationResult) (int, ValidationResult) {

	log2.Debug.Printf("start to calculate conclusion%#v", results)

	resultmap := map[int32][]float64{}
	rejected := []float64{}
	for _, res := range results {
		worker, err := s.Accounting.GetWorker(res.WorkerId)
		if err == nil {
			if res.IsRejected {
				rejected = append(rejected, worker.Credibility)
				log2.Debug.Printf("calc is rejected by %s", worker.Id)
			} else if !res.IsError {
				log2.Debug.Printf("result is %d", res.Data)
				_, exist := resultmap[res.Data]
				if exist {
					resultmap[res.Data] = append(resultmap[res.Data], worker.Credibility)
				} else {
					resultmap[res.Data] = []float64{worker.Credibility}
				}
			} else {
				log2.Debug.Printf("result is error")
			}
		} else {
			log2.Debug.Printf("failed to get worker account %s %#v", res.WorkerId, err)
		}
	}

	resultData := []int32{}
	credsGroups := [][]float64{}
	for data, res := range resultmap {
		credsGroups = append(credsGroups, res)
		resultData = append(resultData, data)
	}
	if len(rejected) > 0 {
		credsGroups = append(credsGroups, rejected)
	}
	var count int
	var group int

	if len(resultData) >= 2 {
		count = 0
		group = 0
	} else {
		count, group = stolerance.CalcNeedWorkerCountAndBestGroup(s.Accounting.AverageCredibility, credsGroups, s.Accounting.CredibilityThreshould)
	}
	if count < 0 || group < 0 {
		return -1, ValidationResult{}
	} else {
		var valResult ValidationResult
		alreadyGetCount := 0
		if group >= len(resultData) {
			valResult = ValidationResult{
				IsRejected: true,
				IsError:    false,
				Data:       0,
			}
			alreadyGetCount = len(rejected)
		} else {
			valResult = ValidationResult{
				IsRejected: false,
				IsError:    false,
				Data:       resultData[group],
			}
			alreadyGetCount = len(resultmap[resultData[group]])
		}

		half := s.Accounting.GetWorkersCount()/2 + 1
		log2.Debug.Printf("half of workers is %d", half)
		needhalf := half - alreadyGetCount
		if needhalf < 0 {
			needhalf = 0
		}
		if count > needhalf {
			count = needhalf
		}

		return count, valResult

	}

}

//バリデーション結果に基づいて評価値を登録する
func (s *Service) commitReputation(holderId string, results []ValidationResult, conclusion *ValidationResult, validateByBridge bool) {
	if conclusion != nil {
		if conclusion.IsRejected == true {
			s.Accounting.UpdateReputation(holderId, false, validateByBridge)
		} else {
			s.Accounting.UpdateReputation(holderId, true, validateByBridge)
		}
	}
	for _, res := range results {
		confirmed := false
		//TODO: nilチェック
		if conclusion == nil {
			confirmed = false
		} else {
			if res.Data == conclusion.Data && res.IsError == false {
				confirmed = true
			}
			if res.IsRejected == true && res.IsError == false && conclusion.IsRejected == true {
				confirmed = true
			}
		}
		s.Accounting.UpdateReputation(res.WorkerId, confirmed, validateByBridge)
	}
}

//ValidatableCodeを検証
func (s *Service) ValidateCode(ctx context.Context, picknum int, holderId string, vCode *pb.ValidatableCode) ([]ValidationResult, *ValidationResult, error) {

	unavailable := []string{}
	waitlist := []string{}
	results := []ValidationResult{}
	var conclusion *ValidationResult

	errChan := make(chan error)                //エラー送信用チャネルs
	resultChan := make(chan *ValidationResult) //バリデーション結果送信チャネル

	nextpick := picknum

	for {
		if nextpick > 0 {
			log2.Debug.Printf("start to pick %d validators", nextpick)
			//ワーカー取得
			pickedWorkers, err := s.Accounting.SelectValidationWorkers(nextpick, unavailable)

			if err == nil {
				log2.Debug.Printf("Success to pick worker %d", len(pickedWorkers))
			} else {
				log2.Err.Printf("Failed to pick worker %#v", err)
			}

			if err == nil {
				nextpick = 0
				pickedIDs := []string{}
				badreputations := []int32{}
				for _, worker := range pickedWorkers {
					pickedIDs = append(pickedIDs, worker.Id)
					if worker.IsBad {
						log2.Debug.Printf("bad woker detected")
						badreputations = append(badreputations, int32(worker.GoodWorkCount))
					}
				}
				unavailable = append(unavailable, pickedIDs...)
				waitlist = append(waitlist, pickedIDs...)

				//各ワーカにValidationリクエスト送信
				for _, worker := range pickedWorkers {
					if s.WithoutConnectRemoteForTest == false {
						go func(target *accounting.Worker) {
							validationResult := s.validateCodeRemote(target, vCode, badreputations)
							resultChan <- validationResult
						}(worker)
					} else {
						go func(target *accounting.Worker) {
							resultChan <- &ValidationResult{WorkerId: target.Id, Data: 10, IsRejected: false, IsError: false}
						}(worker)
					}
				}
			} else {
				log2.Debug.Printf("sending error to errchannel")
				go func() {
					errChan <- err
				}()
			}
		}

		select {
		case res := <-resultChan:
			log2.Debug.Printf("get validation result DATA[%d] FROM[%s] REJECT[%v]", res.Data, res.WorkerId, res.IsRejected)
			tmp := []string{}
			s.Lock()
			for _, id := range waitlist {
				if id != res.WorkerId {
					tmp = append(tmp, id)
				}
			}
			waitlist = tmp
			s.Unlock()
			log2.Debug.Printf("delete ID[%s] from waitlist WAITING[%d]", res.WorkerId, len(waitlist))
			results = append(results, *res)
			var needworker int
			needworker, concval := s.calcNeedAdditionalWorkerAndResult(picknum, results)
			log2.Debug.Printf("need %d more result", needworker)
			conclusion = &concval
			if needworker > 0 {
				conclusion = nil
				nextpick = needworker - len(waitlist)
				log2.Debug.Printf("nextpick is %d", nextpick)
				if nextpick < 0 {
					nextpick = 0
				}
			} else if needworker < 0 {
				return results, nil, ErrFailedToCalcNeedWorker
			} else {
				return results, conclusion, nil
			}
		case <-ctx.Done():
			return results, nil, ctx.Err()
		case err := <-errChan:
			log2.Debug.Printf("Catch error and ending validation")
			return results, nil, err
		default:
		}
	}
}

func (s *Service) DoesValidationNeed(results []ValidationResult, conc ValidationResult) bool {

	for _, res := range results {
		if res.Data != conc.Data {
			return true
		}
	}
	return false
}

func (s *Service) ValidateOnBridge(vcode pb.ValidatableCode) *ValidationResult {
	result := vcode.Data + vcode.Add
	return &ValidationResult{
		Data:       result,
		IsRejected: false,
		IsError:    false,
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

				//利用可能なリクエストが出てくるまでループ
				waitingvreq := []*ValidationRequest{}
				available := false
				var vreq *ValidationRequest = nil
				for available == false {
					vreq = s.dequeueValidationRequest()
					if vreq == nil {
						break
					}
					available = s.IsTheDatapoolAvailable(vreq.DatapoolId)
					if available == false {
						waitingvreq = append(waitingvreq, vreq)
					}
				}

				if vreq != nil && available == true {
					log2.Debug.Printf("Got order to validate")
					//WaitingListに追加
					s.addWaitingList(vreq.DatapoolId)
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
						defer cancel()
						results, conclusion, err := s.ValidateCode(ctx, vreq.Neednum, vreq.HolderId, vreq.ValidatableCode)

						if err != nil {
							log2.Err.Printf("failed to validation %#v", err)
						} else if results != nil {
							validateByBridge := false
							if s.DoesValidationNeed(results, *conclusion) {
								conclusion = s.ValidateOnBridge(*vreq.ValidatableCode)
								s.OnBridgeCount++
								validateByBridge = true
							} else {
								/*conclusion = s.ValidateOnBridge(*vreq.ValidatableCode)
								if s.DoesValidationNeed(results, *conclusion) {
									s.OnBridgeCount++
									validateByBridge = true
								}*/
							}
							//結果から評価値を登録
							fmt.Printf("INFO %s [] END - Validations by remote workers", time.Now())
							s.commitReputation(vreq.HolderId, results, conclusion, validateByBridge)
						}

						//CalcERモードの時、結果をログに出力
						if s.CalcER {
							s.WorksCount += 1
							if results == nil || conclusion == nil {
								s.Lock()
								s.ErrCount += 1
								s.Unlock()
							} else {
								s.NodesSum += len(results)
								s.NodesAvg = float64(s.NodesSum) / float64(s.WorksCount)
								if conclusion.IsRejected {
									s.RejectedCount += 1
								} else if conclusion.IsError {
									s.ErrCount += 1
								} else {
									want := vreq.ValidatableCode.Data + vreq.ValidatableCode.Add
									if want == conclusion.Data {
										s.SuccessCount += 1
									} else {
										s.FailedCount += 1
									}
								}
							}
							log2.TestER.Printf("Validation Complete WORKS[%d] SUC[%d] FAIL[%d] ERR[%d] REJ[%d] NODES[%f] LOSS_B[%d] LOSS_G[%d] BRIDGE_WORKS[%d]", s.WorksCount, s.SuccessCount, s.FailedCount, s.ErrCount, s.RejectedCount, s.NodesAvg, s.Accounting.BadWorkersLoss, s.Accounting.GoodWorkersLoss, s.OnBridgeCount)
						}
						//結果を送信
						if conclusion != nil && conclusion.IsRejected == false {
							s.Datapool.UpdateDatapoolRemote(vreq.DatapoolId, conclusion.Data)
						}
						//WaingListから削除
						s.removeWaitingList(vreq.DatapoolId)
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
