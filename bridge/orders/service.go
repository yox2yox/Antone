package orders

import (
	"context"
	"time"
	"yox2yox/antone/bridge/accounting"
	pb "yox2yox/antone/bridge/pb"
	workerpb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

type OrderResult struct {
	Db         int32
	IsRejected bool
	IsError    bool
}

type Order struct {
	HolderId      string
	CreatedTime   time.Time
	WaitngWorkers []string
	OrderResults  map[string]OrderResult
}

func NewOrder(holderId string, pickedWorker []string) *Order {
	return &Order{
		HolderId:      holderId,
		CreatedTime:   time.Now(),
		WaitngWorkers: pickedWorker,
		OrderResults:  map[string]OrderResult{},
	}
}

type Service struct {
	WaitList   []*Order
	Accounting *accounting.Service
	Debug      bool
}

func NewService(accounting *accounting.Service, debug bool) *Service {
	return &Service{
		WaitList:   []*Order{},
		Accounting: accounting,
		Debug:      debug,
	}
}

func (s *Service) GetOrders() []*Order {
	return s.WaitList
}

func (s *Service) ValidateCode(holderId string, vCode *pb.ValidatableCode) error {

	pickedWorkers, err := s.Accounting.GetValidationWorkers(1)
	if err != nil {
		return err
	}
	pickedIDs := []string{}
	for _, worker := range pickedWorkers {
		pickedIDs = append(pickedIDs, worker.Id)
	}
	order := NewOrder(holderId, pickedIDs)
	s.WaitList = append(s.WaitList, order)

	for _, worker := range pickedWorkers {
		if s.Debug == false {
			go func(target accounting.Worker) {
				conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
				workerClient := workerpb.NewWorkerClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				vCodeWorker := &workerpb.ValidatableCode{Data: vCode.Data, Add: vCode.Add}
				validationResult, err := workerClient.OrderValidation(ctx, vCodeWorker)
				if err != nil {
					order.OrderResults[target.Id] = OrderResult{Db: 0, IsRejected: false, IsError: true}
					return
				} else {
					order.OrderResults[target.Id] = OrderResult{Db: validationResult.Db, IsRejected: false, IsError: true}
					return
				}

			}(worker)
		} else {
			order.OrderResults[worker.Id] = OrderResult{Db: 10, IsRejected: false, IsError: false}
		}
	}
	return nil
}
