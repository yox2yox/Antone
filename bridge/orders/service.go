package orders

import (
	"context"
	"sync"
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
	HolderId       string
	CreatedTime    time.Time
	WaitingWorkers []string
	OrderResults   map[string]OrderResult
}

func NewOrder(holderId string) *Order {
	return &Order{
		HolderId:       holderId,
		CreatedTime:    time.Now(),
		WaitingWorkers: []string{},
		OrderResults:   map[string]OrderResult{},
	}
}

func (o *Order) AddWorkers(workersId []string) {
	o.WaitingWorkers = append(o.WaitingWorkers, workersId...)
}

type Service struct {
	sync.RWMutex
	WaitList                    []*Order //TODO:Remove?
	Accounting                  *accounting.Service
	WithoutConnectRemoteForTest bool
}

func NewService(accounting *accounting.Service, withoutConnectRemoteForTest bool) *Service {
	return &Service{
		WaitList:                    []*Order{},
		Accounting:                  accounting,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
	}
}

//TODO:Remove Order WaitList
func (s *Service) GetOrders() []*Order {
	return s.WaitList
}

func (s *Service) validateCodeRemote(worker *accounting.Worker, vCode *pb.ValidatableCode) *OrderResult {
	conn, err := grpc.Dial(worker.Addr, grpc.WithInsecure())
	workerClient := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	vCodeWorker := &workerpb.ValidatableCode{Data: vCode.Data, Add: vCode.Add}
	validationResult, err := workerClient.OrderValidation(ctx, vCodeWorker)
	if err != nil {
		return &OrderResult{Db: 0, IsRejected: false, IsError: true}
	} else {
		return &OrderResult{Db: validationResult.Pool, IsRejected: false, IsError: false}
	}
}

//ValidatableCodeを検証
func (s *Service) ValidateCode(ctx context.Context, picknum int, holderId string, vCode *pb.ValidatableCode) error {

	order := NewOrder(holderId)
	errChan := make(chan error)
	done := make(chan struct{})
	go func() {
		var mulocal sync.RWMutex
		wg := sync.WaitGroup{}
		doneCount := 0
		for doneCount < picknum {
			//ワーカー取得
			pickedWorkers, err := s.Accounting.SelectValidationWorkers(picknum - doneCount)
			if err != nil {
				errChan <- err
			}
			pickedIDs := []string{}
			for _, worker := range pickedWorkers {
				pickedIDs = append(pickedIDs, worker.Id)
			}
			order.AddWorkers(pickedIDs)

			//各ワーカにValidationリクエスト送信
			for _, worker := range pickedWorkers {
				if s.WithoutConnectRemoteForTest == false {
					wg.Add(1)
					go func(target *accounting.Worker) {
						defer wg.Done()
						orderResult := s.validateCodeRemote(target, vCode)
						if !orderResult.IsError {
							s.Lock()
							order.OrderResults[target.Id] = *orderResult
							s.Unlock()
							mulocal.Lock()
							doneCount += 1
							mulocal.Unlock()
						}
					}(worker)
				} else {
					s.Lock()
					order.OrderResults[worker.Id] = OrderResult{Db: 10, IsRejected: false, IsError: false}
					s.Unlock()
					doneCount += 1
				}
			}
			//全てのワーカへのリクエスト処理が終了するまで待機
			wg.Wait()
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
}
