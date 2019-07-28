package accounting

import (
	"context"
	pb "yox2yox/antone/bridge/pb"
)

type Endpoint struct {
	accounting *Service
}

func NewEndpoint(accounting *Service) *Endpoint {
	return &Endpoint{
		accounting: accounting,
	}
}

//Workerアカウント作成
func (e *Endpoint) SignupWorker(ctx context.Context, req *pb.SignupWorkerRequest) (*pb.WorkerAccount, error) {
	worker, err := e.accounting.CreateNewWorker(req.Id, req.Addr)
	if err != nil {
		return nil, err
	}
	account := &pb.WorkerAccount{
		Id:         worker.Id,
		Balance:    0,
		Reputation: int32(worker.Reputation),
	}
	return account, nil
}

//Clientアカウントを作成
func (e *Endpoint) SignupClient(ctx context.Context, req *pb.SignupClientRequest) (*pb.ClientAccount, error) {
	client, err := e.accounting.CreateNewClient(req.Id)
	if err != nil {
		return nil, err
	}
	account := &pb.ClientAccount{
		Id:      client.Id,
		Balance: int32(client.Balance),
	}
	return account, nil
}

//Workerのアカウント情報を取得
func (e *Endpoint) GetWorkerInfo(ctx context.Context, req *pb.GetWorkerRequest) (*pb.WorkerAccount, error) {
	worker, err := e.accounting.GetWorker(req.Id)
	if err != nil {
		return nil, err
	}
	account := &pb.WorkerAccount{
		Id:         worker.Id,
		Reputation: int32(worker.Reputation),
		Balance:    0,
	}
	return account, nil
}

//Clientのアカウント情報を取得
func (e *Endpoint) GetClientInfo(ctx context.Context, req *pb.GetClientRequest) (*pb.ClientAccount, error) {
	client, err := e.accounting.GetClient(req.Id)
	if err != nil {
		return nil, err
	}
	account := &pb.ClientAccount{
		Id:      client.Id,
		Balance: int32(client.Balance),
	}
	return account, nil
}
