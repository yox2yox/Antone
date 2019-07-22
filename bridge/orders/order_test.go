package orders

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
	"yox2yox/antone/bridge/accounting"
	pb "yox2yox/antone/bridge/pb"

	"google.golang.org/grpc"
)

var (
	tls           = false
	serverAddr    = "127.0.0.1:10000"
	port          = "10000"
	testClientId  = "client0"
	testWorkersId = []string{
		"worker0",
		"worker1",
		"worker2",
		"worker3",
		"worker4",
		"worker5",
		"worker6",
		"worker7",
		"worker8",
		"worker9",
	}
	testWorkerAddr = "addr"
)

func UpServer() (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return nil, nil, err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return grpcServer, lis, nil
}

func TestCreateOrderSuccess(t *testing.T) {
	grpcServer, listen, err := UpServer()
	accounting := accounting.NewService(true)
	orders := NewService(accounting, true)
	endpoint := NewEndpoint(orders, accounting)
	go func() {
		pb.RegisterOrdersServer(grpcServer, endpoint)
		grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()

	for _, workerid := range testWorkersId {
		_, err = accounting.CreateNewWorker(workerid, testWorkerAddr)
		if err != nil {
			t.Fatalf("failed to create worker %#v", err)
		}
	}

	_, err = accounting.RegistarNewDatapoolHolders(testClientId, 1)
	if err != nil {
		t.Fatalf("want no error, but has error %#v", err)
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewOrdersClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orderInfo, err := client.RequestValidatableCode(ctx, &pb.ValidatableCodeRequest{Userid: testClientId, Add: 10})
	t.Logf("%#v aaa", orderInfo)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if orderInfo == nil {
		t.Fatalf("failed test orderinfo is nil")
	}

}

func TestValidateCode(t *testing.T) {
	accounting := accounting.NewService(true)
	for _, workerid := range testWorkersId {
		_, err := accounting.CreateNewWorker(workerid, testWorkerAddr)
		if err != nil {
			t.Fatalf("failed to create new worker %#v", err)
		}
	}

	holder, err := accounting.RegistarNewDatapoolHolders(testClientId, 1)
	if err != nil {
		t.Fatalf("failed to registar holder %#v", err)
	}

	orderService := NewService(accounting, true)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = orderService.ValidateCode(ctx, 5, holder[0].Id, &pb.ValidatableCode{Data: 10, Add: 0})
	if err != nil {
		t.Fatalf("failed to registar validate code %#v", err)
	}

	repcount := 0
	for _, workerid := range testWorkersId {
		rep, err := accounting.GetReputation(workerid)
		if err != nil {
			t.Fatalf("want no error, but has error %#v", err)
		}
		//評価値が1になっている台数を数える
		if rep != 1 {
			repcount += 1
		}
	}
	if repcount != 5 {
		t.Fatalf("want repcount=5, but %#v", repcount)
	}

}

func TestGetValidatableCode(t *testing.T) {
	accounting := accounting.NewService(true)
	_, err := accounting.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("want no error,but has error %#v", err)
	}
	_, err = accounting.CreateNewClient(testClientId)
	if err != nil {
		t.Fatalf("want no error,but has error %#v", err)
	}
	_, err = accounting.RegistarNewDatapoolHolders(testClientId, 1)
	if err != nil {
		t.Fatalf("want no error,but has error %#v", err)
	}
	order := NewService(accounting, true)
	vcode, _, err := order.GetValidatableCode(testClientId, 1)
	if err != nil {
		t.Fatalf("want no error,but has error %#v", err)
	}
	if vcode == nil {
		t.Fatalf("want vcode!=nil,but nil")
	}

}

func TestServiceRunAndStop(t *testing.T) {
	accounting := accounting.NewService(true)
	order := NewService(accounting, true)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		order.Run()
		wg.Done()
	}()
	order.AddValidationRequest(1, testClientId, &pb.ValidatableCode{})
	time.Sleep(3 * time.Second)
	order.Stop()
	wg.Wait()
}
