package datapool

import (
	"errors"
	"sync"
)

type Service struct {
	sync.RWMutex
	dataPool map[string]*int32
	versoin  int64
}

var (
	ErrDataPoolAlreadyExist = errors.New("this user's datapool already exists")
	ErrDataPoolNotExist     = errors.New("this user's datapool does not exist")
)

func NewService() *Service {
	return &Service{
		dataPool: map[string]*int32{},
	}
}

//新規DataPoolの追加
func (s *Service) CreateDataPool(id string, data int32) error {
	s.RLock()
	_, exist := s.dataPool[id]
	s.RUnlock()
	if exist {
		return ErrDataPoolAlreadyExist
	}
	var pool int32 = data
	s.Lock()
	s.dataPool[id] = &pool
	s.versoin = -1
	s.Unlock()
	return nil
}

//Datapoolを削除
func (s *Service) DeleteDataPool(id string) error {
	s.RLock()
	_, exist := s.dataPool[id]
	s.RUnlock()
	if !exist {
		return ErrDataPoolNotExist
	}
	s.Lock()
	delete(s.dataPool, id)
	s.Unlock()
	return nil
}

//idに結びついたdatapoolを取得
func (s *Service) GetDataPool(id string) (int32, int64, error) {
	if !s.ExistDataPool(id) {
		return -1, -1, ErrDataPoolNotExist
	} else {
		s.RLock()
		data := *s.dataPool[id]
		ver := s.versoin
		s.RUnlock()
		return data, ver, nil
	}
}

//idに結びついたデータプールにデータを登録
func (s *Service) SetDataPool(id string, data int32, version int64) error {
	if !s.ExistDataPool(id) {
		return ErrDataPoolNotExist
	}
	s.Lock()
	s.dataPool[id] = &data
	s.versoin = version
	s.Unlock()
	return nil
}

//データプールが存在するか
func (s *Service) ExistDataPool(id string) bool {
	s.RLock()
	_, exist := s.dataPool[id]
	s.RUnlock()
	return exist
}
