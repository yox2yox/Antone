package worker

import (
	"context"
	"errors"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/pkg/stolerance"
	"yox2yox/antone/worker/datapool"
	pb "yox2yox/antone/worker/pb"
)

type Endpoint struct {
	Datapool   *datapool.Service
	Db         map[string]int32
	Id         string
	BadMode    bool
	Reputation int
}

func NewEndpoint(datapool *datapool.Service, badmode bool) *Endpoint {
	return &Endpoint{
		Datapool: datapool,
		Db: map[string]int32{
			"client0": 0,
		},
		BadMode:    badmode,
		Reputation: 0,
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
	log2.Debug.Printf("got bad reputations %#v", validatableCode.Badreputations)

	creds := []float64{}
	for _, rep := range validatableCode.Badreputations {
		wcred := stolerance.CalcWorkerCred(0.4, int(rep))
		creds = append(creds, wcred)
	}
	results := [][]float64{creds}
	log2.Debug.Printf("got bad results %#v", results)
	gcred := stolerance.CalcRGroupCred(0, results)
	log2.Debug.Printf("group cred for bad workers is %f", gcred)
	if e.BadMode && gcred >= 0.9 {
		return &pb.ValidationResult{Pool: -1, Reject: false}, nil
	}
	pool := validatableCode.Data + validatableCode.Add
	e.Reputation += 1
	return &pb.ValidationResult{Pool: pool, Reject: false}, nil
}
