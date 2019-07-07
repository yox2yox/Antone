package worker

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
	pb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

var (
	serverAddr = "127.0.0.1:10000"
	port       = "10000"
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

func TestOrderWorkSuccess(t *testing.T) {
	grpcServer, listen, err := UpServer()
	go func() {
		pb.RegisterWorkerServer(grpcServer, NewEndpoint())
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
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orderResult, err := client.GetValidatableCode(ctx, &pb.ValidatableCodeRequest{Bridgeid: "0", Userid: "0", Add: 10})
	t.Logf("%#v", orderResult)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if orderResult == nil {
		t.Fatalf("failed test orderResult is nil")
	}
}
