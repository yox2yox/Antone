package worker

import (
	"context"
	"errors"
	"yox2yox/antone/worker/datapool"
	pb "yox2yox/antone/worker/pb"
)

type Endpoint struct {
	Datapool *datapool.Service
	Db       map[string]int32
	Id       string
}

func NewEndpoint(datapool *datapool.Service) *Endpoint {
	return &Endpoint{
		Datapool: datapool,
		Db: map[string]int32{
			"client0": 0,
		},
	}
}

var (
	ErrDataPoolAlreadyExist = errors.New("this user's datapool already exists")
	ErrDataPoolNotExist     = errors.New("this user's datapool does not exist")
)

func (e *Endpoint) GetValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	userid := vCodeRequest.Userid
	data, err := e.Datapool.GetDataPool(userid)
	if err == datapool.ErrDataPoolNotExist {
		return nil, ErrDataPoolNotExist
	} else if err != nil {
		return nil, err
	}
	return &pb.ValidatableCode{Data: data, Add: vCodeRequest.Add}, nil
}

func (e *Endpoint) OrderValidation(ctx context.Context, validatableCode *pb.ValidatableCode) (*pb.ValidationResult, error) {
	pool := validatableCode.Data + validatableCode.Add
	return &pb.ValidationResult{Pool: pool, Reject: false}, nil
}

func (e *Endpoint) UpdateDatapool(ctx context.Context, datapoolUpdate *pb.DatapoolUpdate) (*pb.UpdateResult, error) {
	err := e.Datapool.SetDataPool(datapoolUpdate.Userid, datapoolUpdate.Pool)
	if err == datapool.ErrDataPoolNotExist {
		return nil, ErrDataPoolNotExist
	} else if err != nil {
		return nil, err
	}

	return &pb.UpdateResult{}, err
}

func (e *Endpoint) CreateNewDatapool(ctx context.Context, poolInfo *pb.DatapoolInfo) (*pb.CreateDatapoolResult, error) {
	exist := e.Datapool.ExistDataPool(poolInfo.Userid)
	if exist {
		return nil, ErrDataPoolAlreadyExist
	} else {
		err := e.Datapool.CreateNewDataPool(poolInfo.Userid)
		if err != nil {
			return nil, err
		}
	}
	pool, err := e.Datapool.GetDataPool(poolInfo.Userid)
	if err != nil {
		return nil, err
	}
	return &pb.CreateDatapoolResult{Pool: pool}, nil
}
