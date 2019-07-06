package worker

import (
	"context"
	"errors"
	pb "yox2yox/antone/worker/pb"
)

type Endpoint struct {
	db map[string]int32
	id string
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		db: map[string]int32{
			"0": 0,
		},
	}
}

func (e *Endpoint) OrderWork(ctx context.Context, order *pb.Order) (*pb.OrderResult, error) {
	userid := order.Userid
	data, exist := e.db[userid]
	if !exist {
		return nil, errors.New("user's data is not exsist")
	}
	return &pb.OrderResult{Workerid: e.id, Result: data + order.Add}, nil
}

func (e *Endpoint) OrderValidation(ctx context.Context, validatableCode *pb.ValidatableCode) (*pb.ValidationResult, error) {
	result := validatableCode.Data + validatableCode.Add
	//Commit Validation
	return &pb.ValidationResult{}, nil
}
