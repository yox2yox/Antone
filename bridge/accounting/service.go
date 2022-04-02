package accounting

import (
	crand "crypto/rand"
	"errors"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/pkg/stolerance"
)

type Client struct {
	Id      string
	Balance int
}

type Worker struct {
	Addr          string
	Id            string
	Reputation    int
	GoodWorkCount int
	Balance       int
	IsBad         bool
	Holdinds      []string
}

type Service struct {
	sync.RWMutex
	Workers               map[string]*Worker
	Clients               map[string]*Client
	WorkersId             []string
	AverageCredibility    float64
	AverageReputation     float64
	VarianceReputation    float64
	FaultyFraction        float64
	CredibilityThreshould float64
	ReputationResetRate   float64
	//Holders                     map[string][]string
	WithoutConnectRemoteForTest bool
	BadWorkersLoss              int
	GoodWorkersLoss             int
	MaxStake                    int
	MinStake                    int
	StakeLeft                   int
	BlackListing                bool
	StepVoting                  bool
	InitialReputation           int
}

func NewService(withoutConnectRemoteForTest bool, faultyFraction float64, credibilityThreshould float64, reputationResetRate float64, blackListing bool, stepVoting bool, initialReputation int) *Service {
	log2.Debug.Printf("Credibility Threshould is %f", credibilityThreshould)
	return &Service{
		Workers:                     map[string]*Worker{},
		Clients:                     map[string]*Client{},
		WorkersId:                   []string{},
		AverageCredibility:          0.7,
		AverageReputation:           0,
		FaultyFraction:              faultyFraction,
		CredibilityThreshould:       credibilityThreshould,
		ReputationResetRate:         reputationResetRate,
		BadWorkersLoss:              0,
		GoodWorkersLoss:             0,
		MaxStake:                    10000,
		MinStake:                    100,
		StakeLeft:                   0,
		BlackListing:                blackListing,
		StepVoting:                  stepVoting,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
		InitialReputation:           initialReputation,
	}
}

/* Public Functions */
//Worker情報を取得
func (s *Service) GetWorker(workerId string) (Worker, error) {
	workerData := Worker{}
	s.RLock()
	worker, exist := s.Workers[workerId]
	if exist {
		workerData = *worker
	}
	s.RUnlock()
	if !exist {
		return Worker{}, ErrIDNotExist
	}
	return workerData, nil
}

//GetClient はClient情報を取得します
func (s *Service) GetClient(userID string) (Client, error) {
	_, exist := s.Clients[userID]
	if !exist {
		return Client{}, ErrIDNotExist
	}
	return *s.Clients[userID], nil
}

//GetWorkersCount will get workers' count
func (s *Service) GetWorkersCount() int {
	return len(s.Workers)
}

//SelectValidationWorkers exceptionを除くworkerの中からnum台選択して返す
func (s *Service) SelectValidationWorkers(num int, exceptWorkersID []string) ([]*Worker, error) {

	log2.Debug.Printf("start to select %d workers except %#v", num, exceptWorkersID)

	if len(s.Workers)-len(exceptWorkersID) < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())
	selected := []*Worker{}

	for i := 0; i < num; i++ {
		s.RLock()
		pickedID := s.WorkersId[rand.Intn(len(s.WorkersId))]
		s.RUnlock()
		log2.Debug.Printf("worker:%s was picked.", pickedID)
		available := true
		worker, err := s.GetWorker(pickedID)
		if err != nil {
			available = false
		} else {
			log2.Debug.Printf("picked workers balance: %d", worker.Balance)
			if worker.Balance < s.MinStake {
				available = false
			}
		}
		for _, id := range selected {
			if id.Id == pickedID {
				available = false
			}
		}
		for _, id := range exceptWorkersID {
			if id == pickedID {
				available = false
			}
		}
		if available {
			selected = append(selected, s.Workers[pickedID])
			log2.Debug.Printf("worker:%s Address:%s is finaly selected", pickedID, s.Workers[pickedID].Addr)
		} else {
			i--
		}
	}

	return selected, nil
}

//SelectValidationWorkersWithThreshold Thresoldに達するまでワーカーを選出
func (s *Service) SelectValidationWorkersWithThreshold(needAtLeast int, credG [][]float64, targetIndex int, useStepVoting bool, waitlist []string, exceptWorkersID []string) ([]*Worker, error) {

	if len(credG) <= targetIndex {
		return nil, errors.New("index is out of range")
	}

	gotWorkers := []*Worker{}
	for {
		//waitlist内のワーカーのCredibilityを先に取得
		if len(waitlist) > 0 {
			for _, workerid := range waitlist {
				w, err := s.GetWorker(workerid)
				if err == nil {
					var credibility float64 = 0
					if useStepVoting {
						credibility = stolerance.CalcSecondaryWorkerCred(s.FaultyFraction, w.Reputation)
					} else {
						credibility = stolerance.CalcWorkerCred(s.FaultyFraction, w.Reputation)
					}
					credG[targetIndex] = append(credG[targetIndex], credibility)
				}
			}
			waitlist = []string{}
		}

		var gCred float64
		if useStepVoting {
			gCred = stolerance.CalcStepVotingGruopCred(targetIndex, credG)
		} else {
			gCred = stolerance.CalcRGroupCred(targetIndex, credG)
		}
		if gCred >= s.CredibilityThreshould {
			log2.Debug.Printf("Got %d workers", len(gotWorkers))
			if len(gotWorkers) >= needAtLeast {
				return gotWorkers, nil
			}
			need := needAtLeast - len(gotWorkers)
			workers, err := s.SelectValidationWorkers(need, exceptWorkersID)
			if err != nil {
				return nil, err
			}
			gotWorkers = append(gotWorkers, workers...)
			return gotWorkers, nil
		}

		workers, err := s.SelectValidationWorkers(1, exceptWorkersID)
		if err != nil || len(workers) < 1 {
			return nil, err
		}
		var credibility float64
		if s.StepVoting {
			credibility = stolerance.CalcSecondaryWorkerCred(s.FaultyFraction, workers[0].Reputation)
		} else {
			credibility = stolerance.CalcWorkerCred(s.FaultyFraction, workers[0].Reputation)
		}
		gotWorkers = append(gotWorkers, workers[0])
		credG[targetIndex] = append(credG[targetIndex], credibility)
		exceptWorkersID = append(exceptWorkersID, workers[0].Id)

	}
}

//CountBadWorkersReputationLT reputationを下回る評価値を持つ不正ワーカの数を取得
func (s *Service) CountBadWorkersReputationLT(reputation int) int {
	count := 0
	s.RLock()
	for _, worker := range s.Workers {
		if worker.IsBad && worker.Reputation < reputation {
			count++
		}
	}
	s.RUnlock()
	return count
}

/* Setter */
func (s *Service) CreateNewBadWorker(workerId string, Addr string) (*Worker, error) {
	worker, err := s.CreateNewWorker(workerId, Addr)
	if err != nil {
		return nil, err
	} else {
		worker.IsBad = true
		s.StakeLeft += s.MaxStake
		return worker, nil
	}
}

//Workerアカウントを新規作成
func (s *Service) CreateNewWorker(workerId string, Addr string) (*Worker, error) {

	log2.Debug.Printf("creating new worker ID[%s]", workerId)

	s.RLock()
	_, exist := s.Workers[workerId]
	s.RUnlock()
	if exist {
		log2.Err.Printf("failed to create new worker %s", ErrIDAlreadyExists)
		return nil, ErrIDAlreadyExists
	}
	worker := &Worker{
		Addr:          Addr,
		Id:            workerId,
		Reputation:    s.InitialReputation,
		GoodWorkCount: s.InitialReputation,
		Balance:       s.MaxStake,
		IsBad:         false,
		Holdinds:      []string{},
	}
	s.Lock()
	s.Workers[workerId] = worker
	s.WorkersId = append(s.WorkersId, workerId)
	s.Unlock()
	s.calcAverageCredibility()
	log2.Debug.Printf("success to create new worker ID[%s]", workerId)
	return worker, nil
}

//保有しているデータプールの情報をワーカアカウントに記録
func (s *Service) AddHoldingInfo(workerId string, datapoolId string) error {

	log2.Debug.Printf("start adding new Holding datapoolinfo DATAPOOL[%s] to WORKER[%s]", datapoolId, workerId)

	s.RLock()
	_, exist := s.Workers[workerId]
	s.RUnlock()
	if !exist {
		log2.Err.Printf("failed to AddHoldingInfo %s", ErrIDNotExist)
		return ErrIDNotExist
	} else {
		log2.Debug.Printf("sucess to add  DATAPOOL[%s] to WORKER[%s]", datapoolId, workerId)
		s.Lock()
		s.Workers[workerId].Holdinds = append(s.Workers[workerId].Holdinds, datapoolId)
		s.Unlock()
		return nil
	}
}

func (s *Service) RemoveHoldingInfo(workerId string, datapoolid string) error {

	log2.Debug.Printf("start removing DATAPOOL[%s] from WORKER[%s]", datapoolid, workerId)

	s.RLock()
	worker, exist := s.Workers[workerId]
	s.RUnlock()
	if !exist {
		log2.Err.Printf("failed to remove DATAPOOL[%s] from WORKER[%s] %s", datapoolid, workerId, ErrIDNotExist)
		return ErrIDNotExist
	} else {
		holds := []string{}
		for _, hold := range worker.Holdinds {
			if hold != datapoolid {
				holds = append(holds, hold)
			}
		}
		s.Workers[workerId].Holdinds = holds
		log2.Debug.Printf("success to remove DATAPOOL[%s] from WORKER[%s]", datapoolid, workerId)
		return nil
	}
}

//Clientの新規アカウントを作成
func (s *Service) CreateNewClient(userId string) (*Client, error) {

	log2.Debug.Printf("start creating new client ID[%s]", userId)

	s.RLock()
	_, exist := s.Clients[userId]
	s.RUnlock()
	if exist {
		log2.Debug.Printf("failed to create new client ID[%s] %s", userId, ErrIDAlreadyExists)
		return nil, ErrIDAlreadyExists
	}

	client := &Client{
		Id:      userId,
		Balance: 0,
	}
	s.Lock()
	s.Clients[userId] = client
	s.Unlock()
	log2.Debug.Printf("success to create new client ID[%s]", userId)
	return client, nil
}

//評価値を更新
func (s *Service) UpdateReputation(workerId string, confirmed bool, validateByBridge bool) (int, error) {

	log2.Debug.Printf("start updating the reputation of WORKER[%s]", workerId)

	_, exist := s.Workers[workerId]
	if !exist {
		log2.Debug.Printf("failed to update the reputation of WORKER[%s] %s", workerId, ErrIDNotExist)
		return 0, ErrIDNotExist
	}
	rep := 0
	if confirmed {
		//バリデーション成功時の処理

		s.Lock()
		s.Workers[workerId].Reputation++
		s.Workers[workerId].Balance++
		if s.Workers[workerId].IsBad {
			s.StakeLeft++
			s.BadWorkersLoss--
		} else {
			s.GoodWorkersLoss--
		}
		s.Workers[workerId].GoodWorkCount++
		if s.canResetReputaion() {
			s.Workers[workerId].Reputation = 0
		}
		s.Unlock()
		s.calcAverageCredibility()
	} else {
		//バリデーション失敗時の処理
		s.Lock()
		if s.BlackListing {
			s.Workers[workerId].Reputation = 0
			balanceBefore := s.Workers[workerId].Balance
			s.Workers[workerId].Balance = 0
			balanceAfter := s.Workers[workerId].Balance
			sub := balanceBefore - balanceAfter
			if s.Workers[workerId].IsBad {
				log2.Debug.Printf("balance bfore[%d] balance after[%d] sub[%d]", balanceBefore, balanceAfter, sub)
				s.BadWorkersLoss += sub
				s.StakeLeft -= sub
			} else {
				s.GoodWorkersLoss += sub
			}
			s.Workers[workerId].GoodWorkCount = 0
		} else {
			balanceBefore := s.Workers[workerId].Balance
			balanceAfter := 0
			sub := balanceBefore - balanceAfter
			if s.Workers[workerId].IsBad {
				s.BadWorkersLoss += sub
			} else {
				s.GoodWorkersLoss += sub
			}
			if s.canResetReputaion() {
				s.Workers[workerId].Reputation = 0
			}
			s.Workers[workerId].Reputation = 0
			s.Workers[workerId].GoodWorkCount = 0
		}

		rep = s.Workers[workerId].Reputation
		s.Unlock()
		s.calcAverageCredibility()
	}
	log2.Debug.Printf("success to update the reputation of WORKER[%s] to %d", workerId, rep)
	return s.Workers[workerId].Reputation, nil

}

func (s *Service) calcAverageCredibility() float64 {
	sum := 0.0
	sumRep := 0
	s.RLock()
	for _, worker := range s.Workers {
		cred := stolerance.CalcWorkerCred(s.FaultyFraction, worker.Reputation)
		sum += cred
		sumRep += worker.Reputation
	}
	s.RUnlock()
	s.Lock()
	s.AverageCredibility = sum / float64(len(s.Workers))
	s.AverageReputation = float64(sumRep) / float64(len(s.Workers))
	avg := s.AverageReputation
	s.Unlock()
	sumVar := 0.0
	s.RLock()
	for _, worker := range s.Workers {
		cred := float64(worker.Reputation)
		sumVar += (avg - cred) * (avg - cred)
	}
	s.RUnlock()
	s.Lock()
	s.VarianceReputation = math.Sqrt(sumVar / float64(len(s.Workers)))
	s.Unlock()
	return s.AverageCredibility
}

func (s *Service) canResetReputaion() bool {
	n, err := crand.Int(crand.Reader, big.NewInt(1000000))
	if err != nil {
		log2.Err.Printf("failed to get random number")
		return false
	} else {
		resetRate := s.ReputationResetRate
		if n.Cmp(big.NewInt(int64((1-resetRate)*1000000))) == 1 {
			return true
		}
	}
	return false
}
