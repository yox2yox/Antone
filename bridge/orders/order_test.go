package orders

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
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
		pb.RegisterOrdersServer(grpcServer, NewEndpoint())
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
	orderInfo, err := client.CreateOrder(ctx, &pb.WorkRequest{Userid: "0", Add: 10})
	t.Logf("%#v aaa", orderInfo)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if orderInfo == nil {
		t.Fatalf("failed test orderinfo is nil")
	}
}
