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
	Addr       string
	Reputation int
}

type Endpoint struct {
	Orders     *Service
	Accounting *accounting.Service
	PickNum    int
}

func NewEndpoint(orders *Service, accounting *accounting.Service) *Endpoint {
	return &Endpoint{
		Orders:     orders,
		Accounting: accounting,
		PickNum:    1,
	}
}

func (e *Endpoint) RequestValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	if e.Accounting.GetWorkersCount() < e.PickNum {
		return nil, errors.New("There are not enough Wokers")
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
