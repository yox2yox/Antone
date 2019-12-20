package orders

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	ErrFailedToGetWorkers     = errors.New("Failed to get wokrers")
)

type ValidationRequest struct {
	RequestId       int
	DatapoolId      string
	Neednum         int
	HolderId        string
	ValidatableCode *pb.ValidatableCode
}

type ValidationResult struct {
	WorkerId         string
	Data             int32
	IsRejected       bool
	IsError          bool
	IsFistNode       bool
	FirstNodeIsFault bool
}

type Service struct {
	sync.RWMutex
	CommitMutex                 *sync.Mutex
	WaitGroupForValidation      *sync.WaitGroup
	NowValidating               *ValidationRequest
	RequestsBeforeValidation    []*ValidationRequest
	ResultsFromUser             map[int64][]int32
	ResultsBeforeCommit         map[int][]ValidationResult
	ConclusionsBeforeCommit     map[int]*ValidationResult
	RequestsBeforeCommit        map[int]*ValidationRequest
	WaitingTree                 []int
	Accounting                  *accounting.Service
	Datapool                    *datapool.Service
	Running                     bool
	WithoutConnectRemoteForTest bool
	CalcER                      bool
	ExportReputations           int
	StepVoting                  bool
	ErrCount                    int
	FailedCount                 int
	CrossFailCount              int
	SuccessCount                int
	RejectedCount               int
	WorksCount                  int
	LocalWorksCount             int
	OnBridgeCount               int
	NodesAvg                    float64
	NodesSum                    int
	SetWatcher                  bool
	SkipValidation              bool
	SabotagableReputation       int
	WorkerAttackMode            int
	LatestData                  int32
	LatestVersion               int64
	NextRequestId               int
	stopChan                    chan struct{}
}

func NewService(accounting *accounting.Service, datapool *datapool.Service, withoutConnectRemoteForTest bool, calcER bool, setWatcher bool, stepVoting bool, skipValidation bool, workerAttackMode int, sabotagableReputation int, exportReputations int) *Service {
	return &Service{
		WaitGroupForValidation:      &sync.WaitGroup{},
		CommitMutex:                 new(sync.Mutex),
		ResultsFromUser:             map[int64][]int32{},
		WaitingTree:                 []int{},
		NowValidating:               nil,
		RequestsBeforeValidation:    []*ValidationRequest{},
		ResultsBeforeCommit:         map[int][]ValidationResult{},
		ConclusionsBeforeCommit:     map[int]*ValidationResult{},
		RequestsBeforeCommit:        map[int]*ValidationRequest{},
		Accounting:                  accounting,
		Datapool:                    datapool,
		Running:                     false,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
		CalcER:                      calcER,
		ExportReputations:           exportReputations,
		ErrCount:                    0,
		SuccessCount:                0,
		FailedCount:                 0,
		RejectedCount:               0,
		OnBridgeCount:               0,
		LocalWorksCount:             0,
		LatestData:                  0,
		NextRequestId:               0,
		SetWatcher:                  setWatcher,
		StepVoting:                  stepVoting,
		SkipValidation:              skipValidation,
		WorkerAttackMode:            workerAttackMode,
		SabotagableReputation:       sabotagableReputation,
		stopChan:                    make(chan struct{}),
	}
}

func (s *Service) GetWaitingTreeCount() int {
	return len(s.WaitingTree)
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
	vcodeout := &pb.ValidatableCode{Data: vcode.Data, Add: vcode.Add, Version: vcode.Version}
	return vcodeout, nil

}

//ValidatableCodeを取得する
func (s *Service) GetValidatableCode(ctx context.Context, datapoolId string, add int32) (*pb.ValidatableCode, string, int, map[int64]int32, error) {

	if s.SkipValidation == false {
		//データプールが利用可能になるまで待機
		s.WaitGroupForValidation.Wait()
	}
	s.WaitGroupForValidation.Add(1)
	s.Lock()
	myRequestID := s.NextRequestId
	s.ResultsFromUser[int64(myRequestID)] = []int32{}
	s.WaitingTree = append(s.WaitingTree, myRequestID)
	s.NextRequestId++
	s.Unlock()

	holder, err := s.Datapool.SelectDataPoolHolder(datapoolId, []string{})
	if err != nil {
		return nil, "", -1, nil, err
	}
	var vcode *pb.ValidatableCode = nil
	if s.WithoutConnectRemoteForTest {
		vcode = &pb.ValidatableCode{
			Data: 0,
			Add:  add,
		}
		return vcode, holder.Id, myRequestID, nil, nil
	} else {
		vcode, err = s.getValidatableCodeRemote(*holder, datapoolId, add)
		if err != nil {
			return nil, holder.Id, myRequestID, nil, err
		}
		if vcode == nil {
			return nil, holder.Id, myRequestID, nil, ErrGottenDataIsNil
		} else if s.SkipValidation {
			data := vcode.Data
			resultsMap := map[int64]int32{}
			log2.Debug.Printf("Data is %d version[%d]", data, vcode.Version)
			for reqID := vcode.Version + 1; reqID < int64(myRequestID); reqID++ {
				log2.Debug.Printf("Request %d detected", reqID)
				if reqID < int64(myRequestID) && reqID > vcode.Version {
					data++
					resultsMap[reqID] = data
					log2.Debug.Printf("data was incremented %d", data)
					s.LocalWorksCount++
				}
			}
			log2.Debug.Printf("request[%d] vcode is %d + 1", myRequestID, data)
			vcode = &pb.ValidatableCode{Data: data, Add: add}
			return vcode, "", myRequestID, resultsMap, nil
		} else {
			log2.Debug.Printf("request[%d] vcode is %d + 1", myRequestID, vcode.Data)
			return vcode, holder.Id, myRequestID, nil, nil
		}
	}
}

func (s *Service) validateCodeRemote(worker *accounting.Worker, vCode *pb.ValidatableCode, badreputations []int32, firstNodeIsFault bool) *ValidationResult {
	conn, err := grpc.Dial(worker.Addr, grpc.WithInsecure())
	if err != nil {
		log2.Err.Printf("failed to connect to remote %#v", err)
	}
	defer conn.Close()
	workerClient := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	unsabotagable := s.Accounting.CountBadWorkersReputationLT(s.SabotagableReputation)
	vCodeWorker := &workerpb.ValidatableCode{Data: vCode.Data, Add: vCode.Add, Reputation: int32(worker.Reputation), Badreputations: badreputations, Threshould: float32(s.Accounting.CredibilityThreshould), Resetrate: float32(s.Accounting.ReputationResetRate), FirstNodeIsfault: firstNodeIsFault, FaultyFraction: s.Accounting.FaultyFraction, CountUnstabotagable: int32(unsabotagable)}
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
	count := len(s.RequestsBeforeValidation)
	s.RUnlock()
	if count <= 0 {
		return nil
	}
	s.Lock()
	tmp := s.RequestsBeforeValidation[0]
	if len(s.RequestsBeforeValidation) > 1 {
		s.RequestsBeforeValidation = s.RequestsBeforeValidation[1:]
	} else {
		s.RequestsBeforeValidation = []*ValidationRequest{}
	}
	s.Unlock()
	return tmp
}

func (s *Service) AddValidationRequest(requestID int, datapoolId string, needNum int, holderId string, vcode *pb.ValidatableCode, resultsMap map[int64]int32) {
	fmt.Printf("DEBUG %s [] A ValidationRequest is added to orders.service\n", time.Now().String())
	vreq := &ValidationRequest{DatapoolId: datapoolId, Neednum: needNum, HolderId: holderId, ValidatableCode: vcode, RequestId: requestID}
	s.Lock()
	s.RequestsBeforeValidation = append(s.RequestsBeforeValidation, vreq)
	s.Unlock()
	if resultsMap != nil {
		for key, res := range resultsMap {
			s.RLock()
			_, exist := s.ResultsFromUser[key]
			s.RUnlock()
			if exist {
				s.Lock()
				s.ResultsFromUser[key] = append(s.ResultsFromUser[key], res)
				s.Unlock()
			}
		}
	}
	s.addRequestToTree(vreq)
	fmt.Printf("DEBUG %s [] Waiting RequestsBeforeValidation count is %d\n", time.Now().String(), len(s.RequestsBeforeValidation))
}

func (s *Service) getAncestorsFromTree(id int) []*ValidationRequest {
	output := []*ValidationRequest{}
	for _, reqID := range s.WaitingTree {
		if req, exist := s.RequestsBeforeCommit[reqID]; exist && reqID < id {
			output = append(output, req)
		}
	}
	return output
}

func (s *Service) deleteRequestFromTree(id int) {

	count := len(s.WaitingTree)

	if count <= 0 {
		return
	}
	s.Lock()
	newTree := []int{}
	for _, reqID := range s.WaitingTree {
		if reqID != id {
			newTree = append(newTree, reqID)
		}
	}
	s.WaitingTree = newTree
	if _, exist := s.ResultsBeforeCommit[id]; exist {
		delete(s.ResultsBeforeCommit, id)
	}
	if _, exist := s.ConclusionsBeforeCommit[id]; exist {
		delete(s.ConclusionsBeforeCommit, id)
	}
	s.Unlock()

}

func (s *Service) addRequestToTree(vreq *ValidationRequest) {
	s.Lock()
	s.RequestsBeforeCommit[vreq.RequestId] = vreq
	s.Unlock()
}

func (s *Service) commitConclusionToTree(id int, results []ValidationResult, conclusion *ValidationResult) {
	s.CommitMutex.Lock()
	s.Lock()
	s.ResultsBeforeCommit[id] = results
	s.ConclusionsBeforeCommit[id] = conclusion
	s.Unlock()
	log2.Debug.Printf("waiting tree is %#v", s.WaitingTree)
	concludedIDs := []int{}
	wResults := map[int][]ValidationResult{}
	wConclusions := map[int]*ValidationResult{}
	wRequests := map[int]*ValidationRequest{}

	for _, id := range s.WaitingTree {
		s.RLock()
		res, exist := s.ResultsBeforeCommit[id]
		conc, existConc := s.ConclusionsBeforeCommit[id]
		req, existReq := s.RequestsBeforeCommit[id]
		s.RUnlock()
		if exist && existReq && existConc {
			wResults[id] = results
			wConclusions[id] = conc
			wRequests[id] = req
			concludedIDs = append(concludedIDs, id)
			s.commitConclustion(req, res, conc)
			s.deleteRequestFromTree(id)
		} else {
			break
		}
	}
	s.CommitMutex.Unlock()
}

func (s *Service) commitConclustion(vreq *ValidationRequest, results []ValidationResult, conclusion *ValidationResult) {
	validateByBridge := false
	s.RLock()
	resFromUser, exist := s.ResultsFromUser[int64(vreq.RequestId)]
	s.RUnlock()
	if s.DoesValidationNeed(results, *conclusion) && s.SetWatcher {
		conclusion = s.ValidateOnBridge(*vreq.ValidatableCode)
		s.OnBridgeCount++
		validateByBridge = true
	} else if s.SkipValidation && conclusion.FirstNodeIsFault == false {
		conclusion = s.ValidateOnBridge(*vreq.ValidatableCode)
		if s.DoesValidationNeed(results, *conclusion) {
			s.OnBridgeCount++
			validateByBridge = true
		}
	} else if exist {
		for _, data := range resFromUser {
			if conclusion.Data != data {
				conclusion = s.ValidateOnBridge(*vreq.ValidatableCode)
				validateByBridge = true
				break
			}
		}
		s.Lock()
		s.ResultsFromUser[int64(vreq.RequestId)] = nil
		s.Unlock()
	}
	//結果から評価値を登録
	fmt.Printf("INFO %s [] END - Validations by remote workers", time.Now())
	s.commitReputation(vreq.HolderId, results, conclusion, validateByBridge)

	//結果を送信
	if conclusion != nil && conclusion.IsRejected == false {
		res := s.ValidateOnBridge(*vreq.ValidatableCode)
		if res.Data != conclusion.Data {
			s.CrossFailCount++
		}
		s.outputFailureRate(results, conclusion, vreq)
		s.Datapool.UpdateDatapoolRemote(vreq.DatapoolId, vreq.RequestId, conclusion.Data)
	}
}

//バリデーション結果に基づいて評価値を登録する
func (s *Service) commitReputation(holderId string, results []ValidationResult, conclusion *ValidationResult, validateByBridge bool) {
	if conclusion != nil {
		if conclusion.IsRejected == false && conclusion.IsError == false {
			s.LatestData = conclusion.Data
		}
		//データホルダーの評価値を登録
		/*if conclusion.IsRejected == true {
			s.Accounting.UpdateReputation(holderId, false, validateByBridge)
		} else {
			s.Accounting.UpdateReputation(holderId, true, validateByBridge)
		}*/
	}
	for _, res := range results {
		//ワーカの評価値を登録
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
		if confirmed == false {
			s.Accounting.UpdateReputation(res.WorkerId, confirmed, validateByBridge)
		} else if s.StepVoting == false || res.IsFistNode {
			s.Accounting.UpdateReputation(res.WorkerId, confirmed, validateByBridge)
		}
	}
}

func (s *Service) ValidationResultsToCredGroups(results []ValidationResult) ([][]float64, []int32, error) {
	credG := [][]float64{}

	resultmap := map[int32][]float64{}
	rejected := []float64{}
	for _, res := range results {
		worker, err := s.Accounting.GetWorker(res.WorkerId)
		if err == nil {
			var credibility float64 = 0
			if s.StepVoting {
				if res.IsFistNode {
					credibility = stolerance.CalcWorkerCred(s.Accounting.FaultyFraction, worker.Reputation)
				} else {
					credibility = stolerance.CalcSecondaryWorkerCred(s.Accounting.FaultyFraction, worker.Reputation)
				}
			} else {
				credibility = stolerance.CalcWorkerCred(s.Accounting.FaultyFraction, worker.Reputation)
			}
			if res.IsRejected {
				rejected = append(rejected, credibility)
				log2.Debug.Printf("calc is rejected by %s", worker.Id)
			} else if !res.IsError {
				log2.Debug.Printf("result is %d", res.Data)
				_, exist := resultmap[res.Data]
				if exist {
					resultmap[res.Data] = append(resultmap[res.Data], credibility)
				} else {
					resultmap[res.Data] = []float64{credibility}
				}
			} else {
				log2.Debug.Printf("result is error")
			}
		} else {
			log2.Debug.Printf("failed to get worker account %s %#v", res.WorkerId, err)
		}
	}

	resultData := []int32{}
	for data, res := range resultmap {
		credG = append(credG, res)
		resultData = append(resultData, data)
	}
	if len(rejected) > 0 {
		credG = append(credG, rejected)
	}

	return credG, resultData, nil

}

//ValidatableCodeを検証
func (s *Service) ValidateCode(ctx context.Context, picknum int, holderId string, vCode *pb.ValidatableCode) ([]ValidationResult, *ValidationResult, error) {

	unavailable := []string{}
	waitlist := []string{}
	results := []ValidationResult{}

	errChan := make(chan error)                //エラー送信用チャネルs
	resultChan := make(chan *ValidationResult) //バリデーション結果送信チャネル
	log2.Debug.Printf("validation start with %d validators", picknum)
	var nextpicks []*accounting.Worker
	var firstPick *accounting.Worker
	secondPicked := false
	firstNodeIsFault := false
	var err error
	if s.StepVoting {
		nextpicks, err = s.Accounting.SelectValidationWorkers(1, []string{})
		if err != nil {
			log2.Debug.Printf("first pick failed %#v", err)
			return []ValidationResult{}, nil, err
		}
		if len(nextpicks) < 1 {
			log2.Debug.Printf("first pick failed")
			return []ValidationResult{}, nil, ErrFailedToGetWorkers
		} else {
			log2.Debug.Printf("first pick finished")
			firstPick = nextpicks[0]
		}
	} else {
		nextpicks, err = s.Accounting.SelectValidationWorkersWithThreshold(picknum, [][]float64{[]float64{}}, 0, false, []string{}, unavailable)
		if err != nil {
			return []ValidationResult{}, nil, err
		}
	}
	err = nil
	badreputations := []int32{}
	for {
		if len(nextpicks) > 0 {

			if err == nil {
				pickedIDs := []string{}
				for _, worker := range nextpicks {
					pickedIDs = append(pickedIDs, worker.Id)
					if worker.IsBad {
						log2.Debug.Printf("bad woker detected")
						badreputations = append(badreputations, int32(worker.GoodWorkCount))
					}
				}
				unavailable = append(unavailable, pickedIDs...)
				waitlist = append(waitlist, pickedIDs...)

				//各ワーカにValidationリクエスト送信
				for _, worker := range nextpicks {
					if s.WithoutConnectRemoteForTest == false {
						log2.Debug.Printf("Send Validation Request to %s firstNodeIsFault[%#v]", worker.Addr, firstNodeIsFault)
						go func(target *accounting.Worker) {
							validationResult := s.validateCodeRemote(target, vCode, badreputations, firstNodeIsFault)
							resultChan <- validationResult
						}(worker)
					} else {
						go func(target *accounting.Worker) {
							resultChan <- &ValidationResult{WorkerId: target.Id, Data: 10, IsRejected: false, IsError: false}
						}(worker)
					}
				}
				nextpicks = []*accounting.Worker{}
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
			//結果を受け取ったワーカーをwaitlistから削除
			for _, id := range waitlist {
				if id != res.WorkerId {
					tmp = append(tmp, id)
				}
			}
			waitlist = tmp
			s.Unlock()
			log2.Debug.Printf("delete ID[%s] from waitlist WAITING[%d]", res.WorkerId, len(waitlist))
			if s.StepVoting && res.WorkerId == firstPick.Id {
				res.IsFistNode = true
				if res.Data == -1 {
					s.Lock()
					firstNodeIsFault = true
					s.Unlock()
					log2.Debug.Printf("first BadNode sabotaged")
				}
			}
			results = append(results, *res)

			credG, resultData, err := s.ValidationResultsToCredGroups(results)
			if err != nil {
				log2.Err.Printf("Failed to calc results credibility groups %#v", err)
				return results, nil, err
			}
			log2.Debug.Printf("Got Credibility Groups is %#v", credG)
			maxGroup := stolerance.CalcBestGroup(credG, s.Accounting.CredibilityThreshould)
			if s.StepVoting && secondPicked == false {
				nextpicks, err = s.Accounting.SelectValidationWorkersWithThreshold(1, credG, maxGroup, s.StepVoting, waitlist, unavailable)
				secondPicked = true
			} else {
				nextpicks, err = s.Accounting.SelectValidationWorkersWithThreshold(0, credG, maxGroup, s.StepVoting, waitlist, unavailable)
			}
			if err != nil {
				log2.Err.Printf("Failed to calc Next Workers %#v", err) //error happened
				return results, nil, err
			}

			//nextpicksもwaitlistも空になったら終了
			if len(nextpicks) <= 0 && len(waitlist) <= 0 {
				conclusion := &ValidationResult{WorkerId: "", Data: resultData[maxGroup], IsRejected: false, IsError: false, FirstNodeIsFault: true}
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

func (s *Service) outputFailureRate(results []ValidationResult, conclusion *ValidationResult, vreq *ValidationRequest) {
	s.WorksCount++
	if results == nil || conclusion == nil {
		s.Lock()
		s.ErrCount++
		s.Unlock()
	} else {
		s.NodesSum += len(results)
		s.NodesAvg = float64(s.NodesSum) / float64(s.WorksCount)
		if conclusion.IsRejected {
			s.RejectedCount++
		} else if conclusion.IsError {
			s.ErrCount++
		} else {
			want := vreq.ValidatableCode.Data + vreq.ValidatableCode.Add
			if want == conclusion.Data {
				s.SuccessCount++
			} else {
				s.FailedCount++
			}
		}
	}
	reputationsOutput := ""
	for i := 10001; i < 10001+s.ExportReputations; i++ {
		w, err := s.Accounting.GetWorker("localhost:" + strconv.Itoa(i))
		if err == nil {
			reputationsOutput += strconv.Itoa(w.Reputation) + " "
		}
	}
	if reputationsOutput != "" {
		log2.Export.Printf("%d %s", s.WorksCount, reputationsOutput)
	}
	log2.TestER.Printf("Validation Complete WORKS[%d] SUC[%d] FAIL[%d] ERR[%d] REJ[%d] NODES[%d] NODES_AVG[%f] LOSS_B[%d] LEFT_B[%d] LOSS_G[%d] BRIDGE_WORKS[%d] AVG_REP[%f] VAR_REP[%f] Data[%d] REQ_ID[%d] LOCAL_WORKS[%d]", s.WorksCount, s.SuccessCount, s.FailedCount, s.ErrCount, s.RejectedCount, len(results), s.NodesAvg, s.Accounting.BadWorkersLoss, s.Accounting.StakeLeft, s.Accounting.GoodWorkersLoss, s.OnBridgeCount, s.Accounting.AverageReputation, s.Accounting.VarianceReputation, conclusion.Data, vreq.RequestId, s.LocalWorksCount)
}

func (s *Service) Run() {
	for {
		select {
		case <-s.stopChan:
			fmt.Printf("INFO %s [] OrdersServer has been stoped\n", time.Now())
			return
		default:
			if len(s.RequestsBeforeValidation) > 0 && (s.NowValidating == nil || s.SkipValidation) {
				vreq := s.dequeueValidationRequest()
				s.NowValidating = vreq
				if vreq != nil {
					log2.Debug.Printf("Got order to validate")
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
						defer cancel()
						results, conclusion, err := s.ValidateCode(ctx, vreq.Neednum, vreq.HolderId, vreq.ValidatableCode)
						if s.LatestVersion < int64(vreq.RequestId) {
							if conclusion != nil {
								s.LatestData = conclusion.Data //segmentaion
							}
							s.LatestVersion = int64(vreq.RequestId)
						}
						if err != nil {
							log2.Err.Printf("failed to validation %#v", err)
						} else if results != nil {
							s.commitConclusionToTree(vreq.RequestId, results, conclusion)
						}
						s.WaitGroupForValidation.Done()
						s.Lock()
						s.NowValidating = nil
						s.Unlock()
					}()
				}
			}
		}
	}
}

func (s *Service) Stop() {
	close(s.stopChan)
	acc := s.Accounting
	data, _, _ := s.Datapool.FetcheDatapoolFromRemote("client30")
	s.WaitGroupForValidation.Wait()
	log2.Export.Printf("%d %f %f %f %d %t %t %t %t %d %d %d %f %f %f %f %d %d %d %d %d", acc.GetWorkersCount(),
		acc.FaultyFraction, acc.CredibilityThreshould, acc.ReputationResetRate, s.Accounting.InitialReputation,
		s.SetWatcher, s.StepVoting, s.Accounting.BlackListing, s.SkipValidation, s.WorkerAttackMode, s.WorksCount,
		s.FailedCount, float64(s.FailedCount)/float64(s.WorksCount), s.NodesAvg, s.Accounting.AverageReputation,
		s.Accounting.VarianceReputation, s.Accounting.BadWorkersLoss, s.Accounting.GoodWorkersLoss, data,
		s.CrossFailCount, s.LocalWorksCount)
}
