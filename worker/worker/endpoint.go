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
	BadMode  bool
}

func NewEndpoint(datapool *datapool.Service, badmode bool) *Endpoint {
	return &Endpoint{
		Datapool: datapool,
		Db: map[string]int32{
			"client0": 0,
		},
		BadMode: badmode,
	}
}

var (
	ErrDataPoolAlreadyExist = errors.New("this user's datapool already exists")
	ErrDataPoolNotExist     = errors.New("this user's datapool does not exist")
)

func (e *Endpoint) GetValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	datapoolid := vCodeRequest.Datapoolid
	data, err := e.Datapool.GetDataPool(datapoolid)
	if err == datapool.ErrDataPoolNotExist {
		return nil, ErrDataPoolNotExist
	} else if err != nil {
		return nil, err
	}
	return &pb.ValidatableCode{Data: data, Add: vCodeRequest.Add}, nil
}

func (e *Endpoint) OrderValidation(ctx context.Context, validatableCode *pb.ValidatableCode) (*pb.ValidationResult, error) {
	if e.BadMode {
		return &pb.ValidationResult{Pool: -1, Reject: false}, nil
	}
	pool := validatableCode.Data + validatableCode.Add
	return &pb.ValidationResult{Pool: pool, Reject: false}, nil
}
