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
	ErrFailedToComplete          = errors.New("Failed to complete work")
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
func (s *Service) SelectDataPoolHolder(userId string, exceptions []string) (Worker, error) {

	rand.Seed(time.Now().UnixNano())
	s.RLock()
	holders, exist := s.Holders[userId]
	s.RUnlock()
	if exist == false || len(holders) <= 0 {
		return Worker{}, ErrDataPoolHolderNotExist
	}
	if len(holders) < len(exceptions)+1 {
		return Worker{}, ErrWorkersAreNotEnough
	}

	MAXTRY := len(holders) * 5

	holderid := "holder"
	var rtnworker Worker
	continueloop := true
	for i := 0; i < MAXTRY; i++ {
		s.RLock()
		holderid = holders[rand.Intn(len(holders))]
		rtnworker = *s.Workers[holderid]
		s.RUnlock()
		continueloop = false
		for _, exid := range exceptions {
			if holderid == exid {
				continueloop = true
			}
		}
		if continueloop == false {
			return rtnworker, nil
		}
	}
	return Worker{}, ErrFailedToComplete
}

//リモートのデータプールを削除
func (s *Service) DeleteDatapoolOnRemote(userId string, workerId string) error {
	target, exist := s.Workers[workerId]
	if !exist {
		return ErrIDNotExist
	}
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

//ローカルのデータプールを削除
func (s *Service) DeleteDatapoolAndHolderOnLocal(userId string, workerId string) error {
	s.RLock()
	_, exist := s.Holders[userId]
	s.RUnlock()
	if !exist {
		return ErrDataPoolHolderNotExist
	}
	holders := []string{}
	s.RLock()
	readholders := s.Holders[userId]
	s.RUnlock()
	for _, id := range readholders {
		if id == workerId {
			target, exist := s.Workers[workerId]
			if !exist {
				return ErrIDNotExist
			}

			holds := []string{}
			for _, hold := range target.Holdinds {
				if hold != userId {
					holds = append(holds, hold)
				}
			}
			s.Workers[workerId].Holdinds = holds
		} else {
			holders = append(holders, id)
		}
	}
	s.Lock()
	s.Holders[userId] = holders
	s.Unlock()
	return nil
}

//リモートワーカからデータを取得
func (s *Service) FetcheDatapoolFromRemote(userId string) (data int32, failedRemotes []string, err error) {
	maxtry := 20
	exceptions := []string{}
	for i := 0; i < maxtry; i++ {
		holder, err := s.SelectDataPoolHolder(userId, exceptions)
		if err != nil {
			return 0, []string{}, err
		}
		conn, err := grpc.Dial(holder.Addr, grpc.WithInsecure())
		if err != nil {
			exceptions = append(exceptions, holder.Id)
		} else {
			workerClient := workerpb.NewWorkerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			dpinfo := &workerpb.DatapoolInfo{Userid: userId}
			datapool, err := workerClient.GetDatapool(ctx, dpinfo)
			//TODO:データの検証
			if err == nil {
				return datapool.Data, []string{}, nil
			} else {
				exceptions = append(exceptions, holder.Id)
			}
		}
	}
	return 0, []string{}, ErrFailedToComplete

}

//データプールを作成しHolderに登録する
func (s *Service) CreateDatapoolAndSelectHolders(userId string, data int32, num int) ([]Worker, error) {

	if num <= 0 {
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

				datapoolInfo := &workerpb.DatapoolInfo{Userid: userId, Data: data}
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
	//TODO:リモートから返信がなかった場合の処理
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
	_, err := s.CreateDatapoolAndSelectHolders(userId, 0, 1)
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
	mulocal := sync.Mutex{}

	wg := sync.WaitGroup{}

	for _, holder := range holders {
		wg.Add(1)
		go func(target *Worker) {
			defer wg.Done()
			conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
			if err != nil {
				mulocal.Lock()
				failedHolders = append(failedHolders, target.Id)
				mulocal.Unlock()
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
				mulocal.Lock()
				failedHolders = append(failedHolders, target.Id)
				mulocal.Unlock()
				return
			}
		}(holder)
	}

	wg.Wait()

	//エラーを出したデータプールホルダーを交換
	for _, holderid := range failedHolders {
		s.DeleteDatapoolAndHolderOnLocal(userId, holderid)
		if !s.WithoutConnectRemoteForTest {
			s.DeleteDatapoolOnRemote(userId, holderid)
		}
	}
	if len(failedHolders) > 0 {
		s.CreateDatapoolAndSelectHolders(userId, int32(data), len(failedHolders))
	}

	return nil

}
