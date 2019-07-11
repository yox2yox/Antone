package datapool

import (
	"errors"
)

type Service struct {
	dataPool map[string]*int32
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
func (s *Service) CreateNewDataPool(userid string) error {
	//TODO:ミューテーションを考える
	_, exist := s.dataPool[userid]
	if exist {
		return ErrDataPoolAlreadyExist
	}
	var data int32 = 0
	s.dataPool[userid] = &data
	return nil
}

//useridに結びついたdatapoolを取得
func (s *Service) GetDataPool(userid string) (int32, error) {
	if !s.ExistDataPool(userid) {
		return -1, ErrDataPoolNotExist
	} else {
		return *s.dataPool[userid], nil
	}
}

//useridに結びついたデータプールにデータを登録
func (s *Service) SetDataPool(userid string, data int32) error {
	//TODO:ミューテーションを考える
	if !s.ExistDataPool(userid) {
		return ErrDataPoolNotExist
	}
	s.dataPool[userid] = &data
	return nil
}

//データプールが存在するか
func (s *Service) ExistDataPool(userid string) bool {
	_, exist := s.dataPool[userid]
	return exist
}
