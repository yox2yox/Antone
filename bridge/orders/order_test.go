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
	tls        = false
	serverAddr = "127.0.0.1:10000"
	port       = "10000"
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
	go func() {
		pb.RegisterOrdersServer(grpcServer, NewEndpoint(accounting.NewService(true)))
		grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	t.Log("Complete to up server")
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
	orderService := NewService(accounting, true)
	orderService.ValidateCode("holder0", &pb.ValidatableCode{Data: 10, Add: 0})
	if len(orderService.GetOrders()[0].OrderResults) <= 0 {
		t.Fatalf("failed test cannot validate order")
	}
}
