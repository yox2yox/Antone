package orders

import (
	"context"
	"errors"
	"math/rand"
	"time"
	"yox2yox/antone/bridge/accounting"
	"yox2yox/antone/bridge/config"
	pb "yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"
)

type Worker struct {
	Addr       string
	Reputation int
}

type Endpoint struct {
	Config     *config.OrderConfig
	Orders     *Service
	Accounting *accounting.Service
}

func NewEndpoint(config *config.OrderConfig, orders *Service, accounting *accounting.Service) *Endpoint {
	return &Endpoint{
		Orders:     orders,
		Accounting: accounting,
		Config:     config,
	}
}

func (e *Endpoint) RequestValidatableCode(ctx context.Context, vCodeRequest *pb.ValidatableCodeRequest) (*pb.ValidatableCode, error) {
	log2.Debug.Printf("got request for validatable code")

	if e.Accounting.GetWorkersCount() < e.Config.NeedValidationNum {
		log2.Debug.Printf("there are not enough workers")
		return nil, errors.New("There are not enough Wokers")
	}
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//ValidatableCode取得
	vcode, holderId, reqId, resultsMap, err := e.Orders.GetValidatableCode(ctx, vCodeRequest.Datapoolid, vCodeRequest.Add)
	if err != nil {
		log2.Err.Printf("failed to get validatable code %#v", err)
		return nil, err
	}
	log2.Debug.Printf("success to get validatable code %#v", vcode)

	e.Orders.AddValidationRequest(reqId, vCodeRequest.Datapoolid, e.Config.NeedValidationNum, holderId, vcode, resultsMap)
	log2.Debug.Printf("success to add validatable code %#v", vcode)

	//Wait For Validation
	for vCodeRequest.WaitForValidation && e.Orders.GetWaitingTreeCount() > 0 {
		log2.Debug.Printf("waiting validation...[Waiting:%d]", e.Orders.GetWaitingTreeCount())
	}
	return vcode, nil

}

func (e *Endpoint) CommitValidation(ctx context.Context, validationResult *pb.ValidationResult) (*pb.CommitResult, error) {

	return &pb.CommitResult{}, nil
}
