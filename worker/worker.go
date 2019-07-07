package worker

import (
	"context"
	"errors"
	pb "yox2yox/antone/worker/pb"
)

type Endpoint struct {
	Db map[string]int32
	Id string
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		Db: map[string]int32{
			"client0": 0,
		},
	}
}

func (e *Endpoint) GetValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	userid := vCodeRequest.Userid
	data, exist := e.Db[userid]
	if !exist {
		return nil, errors.New("user's data is not exsist")
	}
	return &pb.ValidatableCode{Data: data, Add: vCodeRequest.Add}, nil
}

func (e *Endpoint) OrderValidation(ctx context.Context, validatableCode *pb.ValidatableCode) (*pb.ValidationResult, error) {
	Db := validatableCode.Data + validatableCode.Add
	return &pb.ValidationResult{Db: Db, Reject: false}, nil
}

func (e *Endpoint) UpdateDatabase(ctx context.Context, databaseUpdate *pb.DatabaseUpdate) (*pb.UpdateResult, error) {
	_, exist := e.Db[databaseUpdate.Userid]
	if !exist {
		return nil, errors.New("user's data is not exsist")
	}
	e.Db[databaseUpdate.Userid] = databaseUpdate.Db
	return &pb.UpdateResult{}, nil
}
