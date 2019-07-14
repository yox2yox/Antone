package orders

import (
	"context"
	"fmt"
	"net"
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
	}
	testWorkerAddr = "addr"
)

func UpServer() (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return nil, nil, err
	}
	var opts []grpc.ServerOption
	/*if tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}*/
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

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewOrdersClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orderInfo, err := client.RequestValidatableCode(ctx, &pb.ValidatableCodeRequest{Userid: "0", Add: 10})
	t.Logf("%#v aaa", orderInfo)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if orderInfo == nil {
		t.Fatalf("failed test orderinfo is nil")
	}

}

func TestValidateCodeSuccess(t *testing.T) {
	accounting := accounting.NewService(true)

	_, err := accounting.CreateNewWorker(testWorkersId[0], testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create new worker %#v", err)
	}

	workers, err := accounting.SelectValidationWorkers(1)
	if err != nil {
		t.Fatalf("failed to select worker %#v", err)
	}
	if len(workers) <= 0 || workers[0] == nil {
		t.Fatalf("selected worker's data is broken")
	}

	_, err = accounting.RegistarNewDatapoolHolders(testClientId, len(workers))
	if err != nil {
		t.Fatalf("failed to registar holder %#v", err)
	}

	orderService := NewService(accounting, true)
	err = orderService.ValidateCode(1, workers[0].Id, &pb.ValidatableCode{Data: 10, Add: 0})
	if err != nil {
		t.Fatalf("failed to registar validate code %#v", err)
	}
	waitList := orderService.GetOrders()
	if len(waitList) <= 0 || len(waitList[0].OrderResults) <= 0 {
		t.Fatalf("failed to validate order")
	}
}
