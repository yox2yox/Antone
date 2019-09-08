package accounting

import (
	"errors"
	"math/rand"
	"sync"
	"time"
	"yox2yox/antone/internal/log2"
)

type Client struct {
	Id      string
	Balance int
}

type Worker struct {
	Addr       string
	Id         string
	Reputation int
	Holdinds   []string
}

type Service struct {
	sync.RWMutex
	Workers   map[string]*Worker
	Clients   map[string]*Client
	WorkersId []string
	//Holders                     map[string][]string
	WithoutConnectRemoteForTest bool
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

func NewService(withoutConnectRemoteForTest bool) *Service {
	return &Service{
		Workers:   map[string]*Worker{},
		Clients:   map[string]*Client{},
		WorkersId: []string{},
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
		log2.Debug.Printf("worker:%s is picked", pickedId)
		contain := false
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
		Addr:       Addr,
		Id:         workerId,
		Reputation: 0,
		Holdinds:   []string{},
	}
	s.Lock()
	s.Workers[workerId] = worker
	s.WorkersId = append(s.WorkersId, workerId)
	s.Unlock()
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
func (s *Service) UpdateReputation(workerId string, confirmed bool) (int, error) {

	log2.Debug.Printf("start updating the reputation of WORKER[%s]", workerId)

	_, exist := s.Workers[workerId]
	if !exist {
		log2.Debug.Printf("failed to update the reputation of WORKER[%s] %s", workerId, ErrIDNotExist)
		return 0, ErrIDNotExist
	}
	rep := 0
	if confirmed {
		s.Lock()
		s.Workers[workerId].Reputation += 1
		rep = s.Workers[workerId].Reputation
		s.Unlock()
	} else {
		s.Lock()
		s.Workers[workerId].Reputation -= 1
		rep = s.Workers[workerId].Reputation
		s.Unlock()
	}
	log2.Debug.Printf("success to update the reputation of WORKER[%s] to %d", workerId, rep)
	return s.Workers[workerId].Reputation, nil

}
