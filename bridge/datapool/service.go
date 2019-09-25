package datapool

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"yox2yox/antone/bridge/accounting"
	"yox2yox/antone/internal/log2"
	workerpb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

var (
	ErrNotExist                  = errors.New("doesn't exist")
	ErrDataPoolHolderNotExist    = errors.New("this user's datapool doesn't exist")
	ErrDataPoolAlreadyExists     = errors.New("this user's datapool already exists")
	ErrCreateDataPoolNotComplete = errors.New("failed to complete to create datepool")
	ErrFailedToComplete          = errors.New("Failed to complete work")
	ErrHoldersAreNotEnough       = errors.New("There are not enough holders")
)

type Datapool struct {
	DatapoolId string
	OwnrerId   string
	Holders    []*accounting.Worker
}

type Service struct {
	sync.RWMutex
	accountingService           *accounting.Service
	WithoutConnectRemoteForTest bool
	datapoolList                map[string]*Datapool
}

func NewService(accountinService *accounting.Service, withoutConnectRemoteForTest bool) *Service {
	service := &Service{}

	service.accountingService = accountinService
	service.WithoutConnectRemoteForTest = withoutConnectRemoteForTest
	service.datapoolList = map[string]*Datapool{}

	return service
}

func (s *Service) DatapoolExist(datapoolid string) bool {
	s.RLock()
	_, exist := s.datapoolList[datapoolid]
	s.RUnlock()
	return exist
}

func (s *Service) GetDatapoolHolders(datapoolid string) ([]*accounting.Worker, error) {
	if s.DatapoolExist(datapoolid) == false {
		return nil, ErrNotExist
	}
	return s.datapoolList[datapoolid].Holders, nil
}

//データプールホルダーの中から1台選択して返す
func (s *Service) SelectDataPoolHolder(datapoolId string, exceptions []string) (*accounting.Worker, error) {

	rand.Seed(time.Now().UnixNano())

	s.RLock()
	holders := s.datapoolList[datapoolId].Holders
	if s.DatapoolExist(datapoolId) == false || len(holders) <= 0 {
		return nil, ErrDataPoolHolderNotExist
	}
	if len(holders) < len(exceptions)+1 {
		return nil, ErrHoldersAreNotEnough
	}

	MAXTRY := len(holders) * 5
	s.RUnlock()

	holderid := "holder"
	var rtnworker *accounting.Worker
	continueloop := true
	for i := 0; i < MAXTRY; i++ {
		s.RLock()
		selectindex := rand.Intn(len(s.datapoolList[datapoolId].Holders))
		rtnworker = s.datapoolList[datapoolId].Holders[selectindex]
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
	return nil, ErrFailedToComplete
}

//リモートワーカからデータを取得
func (s *Service) FetcheDatapoolFromRemote(datapoolId string) (data int32, failedRemotes []string, err error) {
	maxtry := 20
	exceptions := []string{}
	for i := 0; i < maxtry; i++ {
		holder, err := s.SelectDataPoolHolder(datapoolId, exceptions)
		if err != nil {
			return 0, []string{}, err
		}
		conn, err := grpc.Dial(holder.Addr, grpc.WithInsecure())
		defer conn.Close()
		if err != nil {
			exceptions = append(exceptions, holder.Id)
		} else {
			dpClient := workerpb.NewDatapoolClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			dpinfo := &workerpb.DatapoolId{Id: datapoolId}
			datapool, err := dpClient.GetDatapool(ctx, dpinfo)
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

//データプールを作成しホルダーを設定
func (s *Service) CreateDatapool(userId string, holdersnum int) (*Datapool, error) {
	index := 0
	for _, exist := s.datapoolList[userId+strconv.Itoa(index)]; exist == true; {
		index += 1
	}
	id := userId + strconv.Itoa(index)
	datapool := &Datapool{}

	datapool.DatapoolId = id
	datapool.OwnrerId = userId
	datapool.Holders = []*accounting.Worker{}

	s.Lock()
	s.datapoolList[id] = datapool
	s.Unlock()
	_, err := s.AddHolders(id, 0, holdersnum)
	if err != nil {
		s.Lock()
		delete(s.datapoolList, id)
		s.Unlock()
		return nil, err
	}

	return datapool, nil
}

//既存のデータプールに対してHolderを追加する
func (s *Service) AddHolders(datapoolid string, data int32, num int) ([]*accounting.Worker, error) {

	log2.Debug.Printf("start adding new holders to DATAPOOL:%s\n", datapoolid)

	if s.DatapoolExist(datapoolid) == false {
		log2.Err.Printf("DATAPOOL:%s does not exist\n", datapoolid)
		return nil, ErrNotExist
	}

	//ホルダーリストをホルダーのIDのリストに変換
	s.RLock()
	alreadyHolds := len(s.datapoolList[datapoolid].Holders)
	holders := s.datapoolList[datapoolid].Holders
	holdersid := []string{}
	for _, holder := range holders {
		holdersid = append(holdersid, holder.Id)
	}
	s.RUnlock()

	rand.Seed(time.Now().UnixNano())

	//新規ホルダーを選出
	newholders, err := s.accountingService.SelectValidationWorkers(num, holdersid)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	for _, worker := range newholders {
		if !s.WithoutConnectRemoteForTest {
			//RemoteワーカにCreateDatapoolリクエスト送信
			wg.Add(1)
			go func(target *accounting.Worker) {
				defer wg.Done()
				log2.Debug.Printf("start sending cretate datapool request to %s\n", target.Addr)
				conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
				defer conn.Close()
				if err != nil {
					log2.Err.Printf("failed to connect to %s\n", target.Addr)
					return
				}
				dpClient := workerpb.NewDatapoolClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				datapoolInfo := &workerpb.DatapoolContent{Id: datapoolid, Data: data}
				_, err = dpClient.CreateDatapool(ctx, datapoolInfo)
				if err != nil {
					log2.Err.Printf("failed to send CreateDatapool request to %s %#v\n", target.Addr, err)
					return
				}
				log2.Debug.Printf("success to create datapool on %s\n", target.Addr)
				s.Lock()
				s.datapoolList[datapoolid].Holders = append(s.datapoolList[datapoolid].Holders, target)
				s.Unlock()
				s.accountingService.AddHoldingInfo(target.Id, datapoolid)
			}(worker)
		} else {
			s.Lock()
			s.datapoolList[datapoolid].Holders = append(s.datapoolList[datapoolid].Holders, worker)
			s.Unlock()
			s.accountingService.AddHoldingInfo(worker.Id, datapoolid)
		}
	}

	if !s.WithoutConnectRemoteForTest {
		wg.Wait()
	}
	s.RLock()
	holders = s.datapoolList[datapoolid].Holders
	getnum := len(holders) - alreadyHolds
	s.RUnlock()
	//TODO:リモートから返信がなかった場合の処理
	if getnum < num {
		log2.Err.Printf("send request is %d but response is %d\n", num, getnum)
		return holders, ErrCreateDataPoolNotComplete
	}

	return holders, nil
}

func (s *Service) UpdateDatapoolRemote(datapoolid string, data int32) error {
	log2.Debug.Println("start updating datapool on remote")
	holders, err := s.GetDatapoolHolders(datapoolid)
	if err != nil {
		return err
	}
	if holders == nil {
		return ErrDataPoolHolderNotExist
	}
	log2.Debug.Printf("success to get %d holders id:[ ", len(holders))
	for _, holder := range holders {
		log2.Debug.Printf("%s, ", holder.Id)
	}
	log2.Debug.Printf("]\n")

	failedHolders := []string{}
	mulocal := sync.Mutex{}

	wg := sync.WaitGroup{}

	for _, holder := range holders {
		wg.Add(1)
		go func(target *accounting.Worker) {
			defer wg.Done()
			log2.Debug.Printf("start access to %s\n", target.Addr)
			conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
			defer conn.Close()
			if err != nil {
				log2.Err.Printf("failed to access to %s\n", target.Addr)
				mulocal.Lock()
				failedHolders = append(failedHolders, target.Id)
				mulocal.Unlock()
				return
			}
			log2.Debug.Printf("success to access to %s\n", target.Addr)
			dpClient := workerpb.NewDatapoolClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			datapoolUpdate := &workerpb.DatapoolContent{
				Id:   datapoolid,
				Data: int32(data),
			}
			log2.Debug.Printf("sending UpdateDatapool request to %s\n", target.Addr)
			_, err = dpClient.UpdateDatapool(ctx, datapoolUpdate)
			if err != nil {
				log2.Err.Printf("failed to request UpdateDatapool to %s\n%s\n", target.Addr, err)
				mulocal.Lock()
				failedHolders = append(failedHolders, target.Id)
				mulocal.Unlock()
				return
			}
			log2.Debug.Printf("success to send UpdateDatapool request to %s\n", target.Addr)
		}(holder)
	}

	wg.Wait()

	//エラーを出したデータプールホルダーを交換
	s.SwapDatapoolHolders(datapoolid, data, failedHolders)

	return nil

}

func (s *Service) SwapDatapoolHolders(datapoolid string, data int32, oldHolders []string) error {
	if s.DatapoolExist(datapoolid) == false {
		return ErrNotExist
	}
	for _, holderid := range oldHolders {
		s.DeleteDatapoolAndHolder(datapoolid, holderid)
	}
	if len(oldHolders) > 0 {
		_, err := s.AddHolders(datapoolid, int32(data), len(oldHolders))
		if err != nil {
			return err
		}
	}
	return nil
}

//ローカルのデータプールを削除
func (s *Service) DeleteDatapoolAndHolder(datapoolid string, workerId string) error {

	if s.DatapoolExist(datapoolid) == false {
		return ErrNotExist
	}

	if !s.WithoutConnectRemoteForTest {
		err := s.deleteDatapoolOnRemote(datapoolid, workerId)
		if err != nil {
			return err
		}
	}

	holders := []*accounting.Worker{}
	s.RLock()
	readholders := s.datapoolList[datapoolid].Holders
	s.RUnlock()

	for _, holder := range readholders {
		if holder.Id == workerId {
			target, err := s.accountingService.GetWorker(workerId)
			if err != nil {
				return err
			}
			s.accountingService.RemoveHoldingInfo(target.Id, datapoolid)
		} else {
			holders = append(holders, holder)
		}
	}
	s.Lock()
	s.datapoolList[datapoolid].Holders = holders
	s.Unlock()
	return nil
}

//リモートのデータプールを削除
func (s *Service) deleteDatapoolOnRemote(datapoolId string, workerId string) error {
	target, err := s.accountingService.GetWorker(workerId)
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	datapoolClient := workerpb.NewDatapoolClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deleteReq := &workerpb.DatapoolId{Id: datapoolId}
	_, err = datapoolClient.DeleteDatapool(ctx, deleteReq)
	if err != nil {
		return err
	}
	return nil
}
