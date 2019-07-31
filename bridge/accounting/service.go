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
	Holdinds   []string
}

type Service struct {
	sync.RWMutex
	Workers                     map[string]*Worker
	Clients                     map[string]*Client
	WorkersId                   []string
	Holders                     map[string][]string
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

func (s *Service) GetDatapoolHolders(userId string) ([]*Worker, error) {
	holders, exist := s.Holders[userId]
	if exist == false {
		return nil, ErrDataPoolHolderNotExist
	}
	workers := []*Worker{}
	for _, workerid := range holders {
		workers = append(workers, s.Workers[workerid])
	}
	return workers, nil
}

//exceptionを除くworkerの中からnum台選択して返す
func (s *Service) SelectValidationWorkers(num int, exception []string) ([]*Worker, error) {
	if len(s.Workers)-len(exception) < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())
	picked := []*Worker{}

	for i := 0; i < num; i++ {
		s.RLock()
		pickedId := s.WorkersId[rand.Intn(len(s.WorkersId))]
		s.RUnlock()
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
		} else {
			i--
		}
	}

	return picked, nil
}

//userIdのデータを保有しているHolderの中から一台選択して返す
func (s *Service) SelectDataPoolHolder(userId string) (*Worker, error) {
	rand.Seed(time.Now().UnixNano())
	s.RLock()
	holders, exist := s.Holders[userId]
	s.RUnlock()
	if exist == false || len(holders) <= 0 {
		return nil, ErrDataPoolHolderNotExist
	}
	s.RLock()
	holderid := holders[rand.Intn(len(holders))]
	rtnworkers := s.Workers[holderid]
	s.RUnlock()
	return rtnworkers, nil
}

//データプールを削除
func (s *Service) DeleteDatapoolAndHolder(userId string, workerId string) error {
	s.RLock()
	_, exist := s.Holders[userId]
	s.RUnlock()
	if !exist {
		return ErrDataPoolHolderNotExist
	}
	for _, id := range s.Holders[userId] {
		if id == workerId {
			target, exist := s.Workers[userId]
			if !exist {
				return ErrIDNotExist
			}

			holds := []string{}
			for _, hold := range target.Holdinds {
				if hold != userId {
					holds = append(holds, hold)
				}
			}
			s.Workers[userId].Holdinds = holds

			conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
			if err != nil {
				return err
			}
			workerClient := workerpb.NewWorkerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			deleteReq := &workerpb.DatapoolInfo{Userid: userId}
			_, err = workerClient.DeleteDatapool(ctx, deleteReq)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return ErrIDNotExist
}

//データプールを作成しHolderに登録する
func (s *Service) CreateDatapoolAndSelectHolders(userId string, num int) ([]Worker, error) {

	//TODO:リモートから返信がなかった場合の処理
	if num < 0 {
		return nil, ErrArgumentIsInvalid
	}

	s.RLock()
	_, exist := s.Holders[userId]
	s.RUnlock()
	if !exist {
		s.Lock()
		s.Holders[userId] = []string{}
		s.Unlock()
	}

	s.RLock()
	lenworkers := len(s.Workers) - len(s.Holders[userId])
	s.RUnlock()
	if lenworkers < num {
		return nil, ErrWorkersAreNotEnough
	}

	rand.Seed(time.Now().UnixNano())

	workers, err := s.SelectValidationWorkers(num, s.Holders[userId])
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	for _, worker := range workers {
		if !s.WithoutConnectRemoteForTest {
			//RemoteワーカにCreateDatapoolリクエスト送信
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
				_, err = workerClient.CreateDatapool(ctx, datapoolInfo)
				if err != nil {
					return
				}
				s.Lock()
				s.Holders[userId] = append(s.Holders[userId], target.Id)
				s.Workers[target.Id].Holdinds = append(s.Workers[target.Id].Holdinds, userId)
				s.Unlock()
			}(worker)
		} else {
			s.Lock()
			s.Holders[userId] = append(s.Holders[userId], worker.Id)
			s.Workers[worker.Id].Holdinds = append(s.Workers[worker.Id].Holdinds, userId)
			s.Unlock()
		}
	}

	if !s.WithoutConnectRemoteForTest {
		wg.Wait()
	}
	holders := []Worker{}
	s.RLock()
	holdersOriginal := s.Holders[userId]
	s.RUnlock()
	if len(holdersOriginal) < num {
		return holders, ErrCreateDataPoolNotComplete
	}
	for _, holder := range holdersOriginal {
		worker, exist := s.Workers[holder]
		if exist {
			holders = append(holders, *worker)
		}
	}

	return holders, nil
}

//Workerアカウントを新規作成
func (s *Service) CreateNewWorker(workerId string, Addr string) (*Worker, error) {
	s.RLock()
	_, exist := s.Workers[workerId]
	s.RUnlock()
	if exist {
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
	return worker, nil
}

//Clientの新規アカウントを作成
func (s *Service) CreateNewClient(userId string) (*Client, error) {
	s.RLock()
	_, exist := s.Clients[userId]
	s.RUnlock()
	if exist {
		return nil, ErrIDAlreadyExists
	}
	_, err := s.CreateDatapoolAndSelectHolders(userId, 1)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Id:      userId,
		Balance: 0,
	}
	s.Lock()
	s.Clients[userId] = client
	s.Unlock()
	return client, nil
}

//評価値を更新
func (s *Service) UpdateReputation(workerId string, confirmed bool) (int, error) {
	_, exist := s.Workers[workerId]
	if !exist {
		return 0, ErrIDNotExist
	}
	if confirmed {
		s.Lock()
		s.Workers[workerId].Reputation += 1
		s.Unlock()
	} else {
		s.Lock()
		s.Workers[workerId].Reputation -= 1
		s.Unlock()
	}
	return s.Workers[workerId].Reputation, nil

}

//リモートワーカのデータプールを更新する
func (s *Service) UpdateDatapoolRemote(userId string, data int) error {
	holders, err := s.GetDatapoolHolders(userId)
	if err != nil {
		return err
	}
	if holders == nil {
		return ErrDataPoolHolderNotExist
	}

	failedHolders := []string{}

	wg := sync.WaitGroup{}

	for _, holder := range holders {
		wg.Add(1)
		go func(target *Worker) {
			defer wg.Done()
			conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
			if err != nil {
				failedHolders = append(failedHolders, target.Id)
				return
			}
			defer conn.Close()
			workerClient := workerpb.NewWorkerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			datapoolUpdate := &workerpb.DatapoolUpdate{
				Userid: userId,
				Pool:   int32(data),
			}
			_, err = workerClient.UpdateDatapool(ctx, datapoolUpdate)
			if err != nil {
				failedHolders = append(failedHolders, target.Id)
				return
			}
		}(holder)
	}

	wg.Wait()

	//TODO:エラーを出したホルダーを新規ホルダーに交換

	return nil

}
