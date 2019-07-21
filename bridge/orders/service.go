package orders

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"yox2yox/antone/bridge/accounting"
	pb "yox2yox/antone/bridge/pb"
	workerpb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

var (
	ErrGottenDataIsNil = errors.New("You got data but it's nil")
)

type ValidationRequest struct {
	Neednum         int
	HolderId        string
	ValidatableCode *pb.ValidatableCode
}

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
	ValidationRequests          []*ValidationRequest
	Accounting                  *accounting.Service
	Running                     bool
	WithoutConnectRemoteForTest bool
	stopChan                    chan struct{}
}

func NewService(accounting *accounting.Service, withoutConnectRemoteForTest bool) *Service {
	return &Service{
		WaitList:                    []*Order{},
		ValidationRequests:          []*ValidationRequest{},
		Accounting:                  accounting,
		Running:                     false,
		WithoutConnectRemoteForTest: withoutConnectRemoteForTest,
		stopChan:                    make(chan struct{}),
	}
}

//TODO:Remove Order WaitList
func (s *Service) GetOrders() []*Order {
	return s.WaitList
}

func (s *Service) getValidatableCodeRemote(holder *accounting.Worker, userId string, add int32) (*pb.ValidatableCode, error) {
	conn, err := grpc.Dial(holder.Addr, grpc.WithInsecure())
	client := workerpb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &workerpb.ValidatableCodeRequest{Userid: userId, Add: add}
	vcode, err := client.GetValidatableCode(ctx, req)
	if err != nil {
		return nil, err
	}
	if vcode == nil {
		return nil, ErrGottenDataIsNil
	}
	vcodeout := &pb.ValidatableCode{Data: vcode.Data, Add: vcode.Add}
	return vcodeout, nil

}

//ValidatableCodeを取得する
func (s *Service) GetValidatableCode(userId string, add int32) (*pb.ValidatableCode, error) {
	holder, err := s.Accounting.SelectDataPoolHolder(userId)
	if err != nil {
		return nil, err
	}
	var vcode *pb.ValidatableCode
	if !s.WithoutConnectRemoteForTest {
		vcode, err = s.getValidatableCodeRemote(holder, userId, add)
		if err != nil {
			return nil, err
		}
	} else {
		vcode = &pb.ValidatableCode{
			Data: 0,
			Add:  add,
		}
	}
	return vcode, nil
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

func (s *Service) dequeueValidationRequest() *ValidationRequest {
	s.Lock()
	tmp := s.ValidationRequests[0]
	s.ValidationRequests = s.ValidationRequests[1:]
	s.Unlock()
	return tmp
}

func (s *Service) AddValidationRequest(needNum int, holderId string, vcode *pb.ValidatableCode) {
	vreq := &ValidationRequest{Neednum: needNum, HolderId: holderId, ValidatableCode: vcode}
	s.Lock()
	s.ValidationRequests = append(s.ValidationRequests, vreq)
	s.Unlock()
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

func (s *Service) Run() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			if len(s.ValidationRequests) > 0 {
				fmt.Println("get item")
				vreq := s.dequeueValidationRequest()
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					s.ValidateCode(ctx, vreq.Neednum, vreq.HolderId, vreq.ValidatableCode)
					//TODO:結果を記録
				}()
			}
		}
	}
}

func (s *Service) Stop() {
	close(s.stopChan)
}
