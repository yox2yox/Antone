package datapool

import (
	"context"
	"yox2yox/antone/worker/pb"
)

type Endpoint struct {
	service *Service
}

func NewEndpoint(service *Service) *Endpoint {
	endpoint := &Endpoint{}
	endpoint.service = service
	return endpoint
}

//Datapoolを取得
func (e *Endpoint) GetDatapool(ctx context.Context, datapoolId *pb.DatapoolId) (*pb.DatapoolContent, error) {
	data, err := e.service.GetDataPool(datapoolId.Id)
	if err != nil {
		return nil, err
	}
	rtnpool := &pb.DatapoolContent{Id: datapoolId.Id, Data: data}
	return rtnpool, nil
}

//Datapoolを更新
func (e *Endpoint) UpdateDatapool(ctx context.Context, datapoolContent *pb.DatapoolContent) (*pb.UpdateResult, error) {
	err := e.service.SetDataPool(datapoolContent.Id, datapoolContent.Data)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateResult{}, err
}

//データプールを新規作成
func (e *Endpoint) CreateDatapool(ctx context.Context, poolContent *pb.DatapoolContent) (*pb.CreateResult, error) {
	exist := e.service.ExistDataPool(poolContent.Id)
	if exist {
		return nil, ErrDataPoolAlreadyExist
	} else {
		err := e.service.CreateDataPool(poolContent.Id, poolContent.Data)
		if err != nil {
			return nil, err
		}
	}
	_, err := e.service.GetDataPool(poolContent.Id)
	if err != nil {
		return nil, err
	}
	return &pb.CreateResult{}, nil
}

//データプールを削除
func (e *Endpoint) DeleteDatapool(ctx context.Context, poolId *pb.DatapoolId) (*pb.DeleteResult, error) {
	exist := e.service.ExistDataPool(poolId.Id)
	if !exist {
		return nil, ErrDataPoolNotExist
	} else {
		err := e.service.DeleteDataPool(poolId.Id)
		if err != nil {
			return nil, err
		}
	}
	return &pb.DeleteResult{}, nil
}
