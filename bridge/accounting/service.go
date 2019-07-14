package accounting

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
	workerpb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
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
	WithoutConnectRemoteForTest bool
}

var (
	ErrIDNotExist             = errors.New("this id doesn't exist")
	ErrIDAlreadyExists        = errors.New("this id already exsits")
	ErrWorkersAreNotEnough    = errors.New("There are not enough workers")
	ErrDataPoolHolderNotExist = errors.New("this user's datapool doesn't exist")
	ErrDataPoolAlreadyExists  = errors.New("this user's datapool already exists")
)

func NewService(withoutConnectRemoteForTest bool) *Service {
	return &Service{
		Workers:                     map[string]*Worker{},
		Clients:                     map[string]*Client{},
		WorkersId:                   []string{},
		Holders:                     map[string][]string{},
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
	}
}

func (s *Service) GetWorkersCount() int {
	return len(s.Workers)
}

//workerの中からnum台選択して返す
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

//userIdのデータを保有しているHolderの中から一台選択して返す
func (s *Service) SelectDataPoolHolder(userId string) (*Worker, error) {
	rand.Seed(time.Now().UnixNano())
	_, exist := s.Holders[userId]
	if exist == false {
		return nil, ErrDataPoolHolderNotExist
	}
	holderid := s.Holders[userId][rand.Intn(len(s.Holders))]
	return s.Workers[holderid], nil
}

//新規データプールを作成しHolderを登録する
func (s *Service) RegistarNewDatapoolHolders(userId string, num int) ([]*Worker, error) {
	if len(s.Workers) < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())

	_, exist := s.Holders[userId]
	if exist {
		return nil, ErrDataPoolAlreadyExists
	}
	s.Holders[userId] = []string{}

	workers, err := s.SelectValidationWorkers(num)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	for _, worker := range workers {
		if !s.WithoutConnectRemoteForTest {
			//RemoteワーカにCreateNewDatapoolリクエスト送信
			wg.Add(1)
			go func(target *Worker) {
				defer wg.Done()
				conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
				if err != nil {
					return
				}
				workerClient := workerpb.NewWorkerClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				datapoolInfo := &workerpb.DatapoolInfo{Userid: userId}
				_, err = workerClient.CreateNewDatapool(ctx, datapoolInfo)
				if err != nil {
					return
				}
				s.Holders[userId] = append(s.Holders[userId], target.Id)
			}(worker)
		} else {
			s.Holders[userId] = append(s.Holders[userId], worker.Id)
		}
	}

	if !s.WithoutConnectRemoteForTest {
		wg.Wait()
	}
	holders := []*Worker{}
	for _, holder := range s.Holders[userId] {
		worker, exist := s.Workers[holder]
		if exist {
			holders = append(holders, worker)
		}
	}

	return holders, nil
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
	s.WorkersId = append(s.WorkersId, workerId)
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
