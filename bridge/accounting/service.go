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
}

var (
	ErrIDNotExist                = errors.New("this id doesn't exist")
	ErrIDAlreadyExists           = errors.New("this id already exsits")
	ErrWorkersAreNotEnough       = errors.New("There are not enough workers")
	ErrDataPoolHolderNotExist    = errors.New("this user's datapool doesn't exist")
	ErrDataPoolAlreadyExists     = errors.New("this user's datapool already exists")
	ErrCreateDataPoolNotComplete = errors.New("failed to complete to create datepool")
	ErrArgumentIsInvalid         = errors.New("an argument is invalid")
	ErrFailedToComplete          = errors.New("Failed to complete work")
)

func NewService(withoutConnectRemoteForTest bool, faultyFraction float64, credibilityThreshould float64, reputationResetRate float64, blackListing bool, stepVoting bool) *Service {
	log2.Debug.Printf("Credibility Threshould is %f", credibilityThreshould)
	return &Service{
		Workers:               map[string]*Worker{},
		Clients:               map[string]*Client{},
		WorkersId:             []string{},
		AverageCredibility:    0.7,
		AverageReputation:     0,
		FaultyFraction:        faultyFraction,
		CredibilityThreshould: credibilityThreshould,
		ReputationResetRate:   reputationResetRate,
		BadWorkersLoss:        0,
		GoodWorkersLoss:       0,
		MaxStake:              10000,
		MinStake:              100,
		StakeLeft:             0,
		BlackListing:          blackListing,
		StepVoting:            stepVoting,
		//Holders:                     map[string][]string{},
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
	}
}

func (s *Service) GetWorker(workerId string) (Worker, error) {
	_, exist := s.Workers[workerId]
	if !exist {
		return Worker{}, ErrIDNotExist
	}
	return *s.Workers[workerId], nil
}

func (s *Service) GetClient(userId string) (Client, error) {
	_, exist := s.Clients[userId]
	if !exist {
		return Client{}, ErrIDNotExist
	}
	return *s.Clients[userId], nil
}

func (s *Service) GetWorkersCount() int {
	return len(s.Workers)
}

//exceptionを除くworkerの中からnum台選択して返す
func (s *Service) SelectValidationWorkers(num int, exception []string) ([]*Worker, error) {

	log2.Debug.Printf("starting select %d workers except %#v", num, exception)

	if len(s.Workers)-len(exception) < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())
	picked := []*Worker{}

	for i := 0; i < num; i++ {
		s.RLock()
		pickedId := s.WorkersId[rand.Intn(len(s.WorkersId))]
		s.RUnlock()
		log2.Debug.Printf("worker:%s is picked.", pickedId)
		contain := false
		worker, err := s.GetWorker(pickedId)
		if err != nil {
			contain = true
		} else {
			log2.Debug.Printf("picked workers balance: %d", worker.Balance)
			if worker.Balance < s.MinStake {
				contain = true
			}
		}
		for _, id := range picked {
			if id.Id == pickedId {
				contain = true
			}
		}
		for _, id := range exception {
			if id == pickedId {
				contain = true
			}
		}
		if contain == false {
			picked = append(picked, s.Workers[pickedId])
			log2.Debug.Printf("worker:%s Address:%s is finaly selected", pickedId, s.Workers[pickedId].Addr)
		} else {
			i--
		}
	}

	return picked, nil
}

//Thresoldに達するまでワーカーを選出
func (s *Service) SelectValidationWorkersWithThreshold(needAtLeast int, credG [][]float64, targetIndex int, stepVoting bool, waitlist []string, exception []string) ([]*Worker, error) {

	if len(credG) <= targetIndex {
		return nil, errors.New("index is out of range")
	}

	gotWorkers := []*Worker{}
	for {
		if len(waitlist) > 0 {
			for _, workerid := range waitlist {
				w, err := s.GetWorker(workerid)
				if err == nil {
					var credibility float64 = 0
					if stepVoting {
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
		if stepVoting {
			gCred = stolerance.CalcStepVotingGruopCred(targetIndex, credG)
		} else {
			gCred = stolerance.CalcRGroupCred(targetIndex, credG)
		}
		if gCred >= s.CredibilityThreshould {
			log2.Debug.Printf("Got %d workers", len(gotWorkers))
			if len(gotWorkers) >= needAtLeast {
				return gotWorkers, nil
			} else {
				need := needAtLeast - len(gotWorkers)
				workers, err := s.SelectValidationWorkers(need, exception)
				if err != nil {
					return nil, err
				} else {
					gotWorkers = append(gotWorkers, workers...)
					return gotWorkers, nil
				}
			}
		} else {
			workers, err := s.SelectValidationWorkers(1, exception)
			if err != nil || len(workers) < 1 {
				return nil, err
			}
			credibility := stolerance.CalcWorkerCred(s.FaultyFraction, workers[0].Reputation)
			gotWorkers = append(gotWorkers, workers[0])
			credG[targetIndex] = append(credG[targetIndex], credibility)
			exception = append(exception, workers[0].Id)
		}
	}
}

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
		Reputation:    0,
		GoodWorkCount: 0,
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

	worker, exist := s.Workers[workerId]
	if !exist {
		log2.Debug.Printf("failed to update the reputation of WORKER[%s] %s", workerId, ErrIDNotExist)
		return 0, ErrIDNotExist
	}
	rep := 0
	if confirmed {
		//バリデーション成功時の処理

		s.Lock()
		s.Workers[workerId].Reputation += 1
		s.Workers[workerId].Balance += 1
		if s.Workers[workerId].IsBad {
			s.StakeLeft += 1
		}
		s.Workers[workerId].GoodWorkCount += 1

		stakeRate := float64(worker.Balance) / float64(s.MaxStake)
		if stakeRate > 1 {
			stakeRate = 1
		}
		rep = s.Workers[workerId].Reputation
		n, err := crand.Int(crand.Reader, big.NewInt(1000000))
		if err != nil {
			log2.Err.Printf("failed to get random number")
		} else {
			resetRate := s.ReputationResetRate
			if n.Cmp(big.NewInt(int64((1-resetRate)*stakeRate*1000000))) > -1 {
				//reset, err := crand.Int(crand.Reader, big.NewInt(100))
				if err != nil {
					log2.Err.Printf("failed to get random number")
					s.Workers[workerId].Reputation = 0
				} else {
					s.Workers[workerId].Reputation = 0
					//s.Workers[workerId].Reputation = int(float64(s.Workers[workerId].Reputation) * (float64(reset.Int64()) / 100))
				}
				rep = s.Workers[workerId].Reputation
			}
		}
		s.Unlock()
		s.calcAverageCredibility()
	} else {
		//バリデーション失敗時の処理
		s.Lock()
		s.Workers[workerId].Reputation = 0
		if s.BlackListing {
			balanceBefore := s.Workers[workerId].Balance
			if validateByBridge {
				s.Workers[workerId].Balance = 0
			} else {
				s.Workers[workerId].Balance = 0
				log2.Err.Printf("Faild to validate On Bridge")
			}
			balanceAfter := s.Workers[workerId].Balance
			sub := balanceBefore - balanceAfter
			if s.Workers[workerId].IsBad {
				log2.Debug.Printf("balance bfore[%d] balance after[%d] sub[%d]", balanceBefore, balanceAfter, sub)
				s.BadWorkersLoss += sub
				s.StakeLeft -= sub
			} else {
				s.GoodWorkersLoss += sub
			}
		}
		s.Workers[workerId].GoodWorkCount = 0
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
	for _, worker := range s.Workers {
		cred := stolerance.CalcWorkerCred(s.FaultyFraction, worker.Reputation)
		sum += cred
		sumRep += worker.Reputation
	}
	s.Lock()
	s.AverageCredibility = sum / float64(len(s.Workers))
	s.AverageReputation = float64(sumRep) / float64(len(s.Workers))
	avg := s.AverageReputation
	s.Unlock()
	sumVar := 0.0
	for _, worker := range s.Workers {
		cred := stolerance.CalcWorkerCred(s.FaultyFraction, worker.Reputation)
		sumVar += (avg - cred) * (avg - cred)
	}
	s.Lock()
	s.VarianceReputation = math.Sqrt(sumVar / float64(len(s.Workers)))
	s.Unlock()
	return s.AverageCredibility
}
