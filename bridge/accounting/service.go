package accounting

import (
	"errors"
	"math/rand"
	"time"
)

type Client struct {
	Id      string
	Balance int
}

type Worker struct {
	Addr       string
	Id         string
	Reputation int
}

type Service struct {
	Workers                     map[string]*Worker
	Clients                     map[string]*Client
	WorkersId                   []string
	Holders                     map[string][]string
	WithoutCommunicationForTest bool
}

var (
	ErrIDAlreadyExists     = errors.New("this id already exsits")
	ErrWorkersAreNotEnough = errors.New("There are not enough workers")
)

func NewService(withoutCommunicationForTest bool) *Service {
	return &Service{
		Workers: map[string]*Worker{
			"worker0": &Worker{
				Addr:       "127.0.0.1:10000",
				Id:         "woker0",
				Reputation: 0,
			},
		},
		Clients: map[string]*Client{},
		WorkersId: []string{
			"worker0",
		},
		Holders: map[string][]string{
			"client0": []string{
				"worker0",
			},
		},
		WithoutCommunicationForTest: withoutCommunicationForTest,
	}
}

func (s *Service) GetWorkersCount() int {
	return len(s.Workers)
}

func (s *Service) SelectValidationWorkers(num int) ([]*Worker, error) {
	if len(s.Workers) < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())
	picked := []*Worker{}
	for i := 0; i < num; i++ {
		pickedId := s.WorkersId[rand.Intn(len(s.Workers))]
		picked = append(picked, s.Workers[pickedId])
	}

	return picked, nil
}

func (s *Service) SelectDBHolder(userId string) *Worker {
	rand.Seed(time.Now().UnixNano())
	holderid := s.Holders[userId][rand.Intn(len(s.Holders))]
	return s.Workers[holderid]
}

//Workerアカウントを新規作成
func (s *Service) CreateNewWorker(workerId string, Addr string) (*Worker, error) {
	_, exist := s.Workers[workerId]
	if exist {
		return nil, ErrIDAlreadyExists
	}
	s.Workers[workerId] = &Worker{
		Addr:       Addr,
		Id:         workerId,
		Reputation: 0,
	}
	return s.Workers[workerId], nil
}

//Clientの新規アカウントを作成
func (s *Service) CreateNewClient(userId string) (*Client, error) {
	_, exist := s.Clients[userId]
	if exist {
		return nil, ErrIDAlreadyExists
	}

	s.Clients[userId] = &Client{
		Id:      userId,
		Balance: 0,
	}
	return s.Clients[userId], nil
}
