package accounting_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
	"yox2yox/antone/bridge/accounting"
	pb "yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"

	"google.golang.org/grpc"
)

var (
	testWorkersId = []string{
		"worker0",
		"worker1",
	}
	testWorkerAddr = "addr"
	testUserId     = "client0"
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

//WorkersCountが正しく機能するか
func TestAccountingService_WorkersCount(t *testing.T) {
	accounting := accounting.NewService(true)
	_, err := accounting.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}
	count := accounting.GetWorkersCount()
	if count != 1 {
		t.Fatalf("failed to count workers\nexpect:1 result:%d", count)
	}
}

func TestAccountingService_SelectValidationWorkers(t *testing.T) {
	accounting := accounting.NewService(true)

	_, err := accounting.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	workers, err := accounting.SelectValidationWorkers(1, []string{})
	if err != nil {
		t.Fatalf("failed to get validation workers%#v", err)
	}
	if len(workers) != 1 {
		t.Fatalf("the number of workers is not expected\nexpected:1 result:%d", len(workers))
	}
	for _, worker := range workers {
		if worker.Id == "" || worker.Addr == "" {
			t.Fatalf("gotten workerdata is broken")
		}
	}

	_, err = accounting.CreateNewWorker(testWorkersId[1], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	workers, err = accounting.SelectValidationWorkers(2, []string{})
	if err != nil {
		t.Fatalf("failed to get validation workers%#v", err)
	}
	if len(workers) != 2 {
		t.Fatalf("the number of workers is not expected\nexpected:2 result:%d", len(workers))
	}
	for _, worker := range workers {
		if worker.Id == "" || worker.Addr == "" {
			t.Fatalf("gotten workerdata is broken")
		}
	}

}

func TestAccountingService_SelectValidationWorkersFail(t *testing.T) {
	accountingS := accounting.NewService(true)

	_, err := accountingS.SelectValidationWorkers(1, []string{})
	if err != accounting.ErrWorkersAreNotEnough {
		t.Fatalf("failed to get NotEnough error%#v", err)
	}

	_, err = accountingS.CreateNewWorker(testWorkersId[1], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	_, err = accountingS.SelectValidationWorkers(2, []string{})
	if err != accounting.ErrWorkersAreNotEnough {
		t.Fatalf("failed to get NotEnough error%#v", err)
	}

}

func TestAccountingService_CreateNewWorkerSuccess(t *testing.T) {
	accounting := accounting.NewService(true)
	worker, err := accounting.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
}

func TestAccountingService_CreateDupulicatedWorkerFail(t *testing.T) {
	accountingS := accounting.NewService(true)
	worker, err := accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if worker == nil {
		t.Fatalf("created worker is nil")
	}
	worker, err = accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != accounting.ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}

func TestAccountingService_CreateNewClientSuccess(t *testing.T) {
	accountingS := accounting.NewService(true)
	_, err := accountingS.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	client, err := accountingS.CreateNewClient(testUserId)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}
	if client == nil {
		t.Fatalf("want client != nil,but nil")
	}
}

func TestAccountingService_CreateDupulicatedClientFail(t *testing.T) {
	accountingS := accounting.NewService(true)
	client, err := accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != nil {
		t.Fatalf("failed to create worker%#v", err)
	}
	if client == nil {
		t.Fatalf("created worker is nil")
	}
	client, err = accountingS.CreateNewWorker("worker:hoge", "hoge")
	if err != accounting.ErrIDAlreadyExists {
		t.Fatalf("couldn't get dupulication error %#v", err)
	}
}

func UpServer(addr string, port string) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return nil, nil, err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return grpcServer, lis, nil
}

func TestEndpoint(t *testing.T) {

	addr := "localhost:10000"
	port := "10000"
	workerId := "worker0"
	clientId := "client0"

	/* make server endpoint */
	accountingS := accounting.NewService(true)
	endpoint := accounting.NewEndpoint(accountingS)
	grpcServer, listen, err := UpServer(addr, port)
	if err != nil {
		t.Fatalf("want no error, but error %#v", err)
	}
	defer listen.Close()
	go func() {
		pb.RegisterAccountingServer(grpcServer, endpoint)
		err := grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("want no error,but error %#v", err)
		}
	}()
	defer grpcServer.Stop()

	/* make client */
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	defer conn.Close()
	client := pb.NewAccountingClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	/* Test Signup Worker */
	reqSignupWorker := &pb.SignupWorkerRequest{
		Id:   workerId,
		Addr: addr,
	}
	workerAccount, err := client.SignupWorker(ctx, reqSignupWorker)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if workerAccount == nil {
		t.Fatalf("want worker account not nil,but nil")
	}
	if workerAccount.Id != workerId {
		t.Fatalf("want workerId == %#v,but %#v", workerId, workerAccount.Id)
	}
	if workerAccount.Reputation != 0 {
		t.Fatalf("want worker reputation is 0,but %#v", workerAccount.Reputation)
	}
	if workerAccount.Balance != 0 {
		t.Fatalf("want worker balance is 0,but %#v", workerAccount.Balance)
	}
	workerlocal, err := accountingS.GetWorker(workerId)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if workerlocal.Id != workerAccount.Id {
		t.Fatalf("want workerlocal id == %#v,but %#v", workerAccount.Id, workerlocal.Id)
	}
	if workerlocal.Reputation != 0 {
		t.Fatalf("want workerlocal reputation == 0,but %#v", workerlocal.Reputation)
	}
	if workerlocal.Addr != addr {
		t.Fatalf("want worker address is %#v,but %#v", addr, workerlocal.Addr)
	}

	/* Test SignupClient */
	reqSignupClient := &pb.SignupClientRequest{
		Id:   clientId,
		Addr: addr,
	}
	clientAccount, err := client.SignupClient(ctx, reqSignupClient)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if clientAccount == nil {
		t.Fatalf("want client account not nil,but nil")
	}
	if clientAccount.Id != clientId {
		t.Fatalf("want client id is %#v,but %#v", clientId, clientAccount.Id)
	}
	if clientAccount.Balance != 0 {
		t.Fatalf("want client balance is 0,but %#v", clientAccount.Balance)
	}
	clientlocal, err := accountingS.GetClient(clientId)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if clientlocal.Id != clientAccount.Id {
		t.Fatalf("want clientlocal id == %#v,but %#v", clientAccount.Id, clientlocal.Id)
	}
	if clientlocal.Balance != 0 {
		t.Fatalf("want clientlocal balance == 0,but %#v", clientlocal.Balance)
	}

	/* Test GetWorkerInfo */
	reqGetWorker := &pb.GetWorkerRequest{
		Id: workerId,
	}
	workerAccount, err = client.GetWorkerInfo(ctx, reqGetWorker)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if workerAccount == nil {
		t.Fatalf("want worker account not nil,but nil")
	}
	if workerAccount.Id != workerId {
		t.Fatalf("want workerId == %#v,but %#v", workerId, workerAccount.Id)
	}
	if workerAccount.Reputation != 0 {
		t.Fatalf("want worker reputation is 0,but %#v", workerAccount.Reputation)
	}
	if workerAccount.Balance != 0 {
		t.Fatalf("want worker balance is 0,but %#v", workerAccount.Balance)
	}
	if workerlocal.Id != workerAccount.Id {
		t.Fatalf("want workerlocal id == %#v,but %#v", workerAccount.Id, workerlocal.Id)
	}
	if workerlocal.Reputation != 0 {
		t.Fatalf("want workerlocal reputation == 0,but %#v", workerlocal.Reputation)
	}
	if workerlocal.Addr != addr {
		t.Fatalf("want worker address is %#v,but %#v", addr, workerlocal.Addr)
	}

	/* Test GetClientInfo */
	reqGetClient := &pb.GetClientRequest{
		Id: clientId,
	}
	clientAccount, err = client.GetClientInfo(ctx, reqGetClient)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}
	if clientAccount == nil {
		t.Fatalf("want client account not nil,but nil")
	}
	if clientAccount.Id != clientId {
		t.Fatalf("want client id is %#v,but %#v", clientId, clientAccount.Id)
	}
	if clientAccount.Balance != 0 {
		t.Fatalf("want client balance is 0,but %#v", clientAccount.Balance)
	}
	if clientlocal.Id != clientAccount.Id {
		t.Fatalf("want clientlocal id == %#v,but %#v", clientAccount.Id, clientlocal.Id)
	}
	if clientlocal.Balance != 0 {
		t.Fatalf("want clientlocal balance == 0,but %#v", clientlocal.Balance)
	}

}
