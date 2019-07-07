package orders

import (
	"context"
	"errors"
	"math/rand"
	"time"
	"yox2yox/antone/bridge/accounting"
	pb "yox2yox/antone/bridge/pb"
)

type Worker struct {
	addr       string
	reputation int
}

type Endpoint struct {
	accounting *accounting.Service
	pickNum    int
}

func NewEndpoint(accounting *accounting.Service) *Endpoint {
	return &Endpoint{
		accounting: accounting,
		pickNum:    1,
	}
}

func (e *Endpoint) RequestValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	if e.accounting.GetWorkersCount() < e.pickNum {
		return nil, errors.New("There is not enough Wokers")
	}
	rand.Seed(time.Now().UnixNano())

	//Send OrderRequest

	return &pb.ValidatableCode{
		Data: 3,
		Add:  vCodeRequest.Add,
	}, nil

}

func (e *Endpoint) CommitValidation(ctx context.Context, validationResult *pb.ValidationResult) (*pb.CommitResult, error) {

	return &pb.CommitResult{}, nil
}
