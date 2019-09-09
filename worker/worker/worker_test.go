package worker_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/worker/datapool"
	pb "yox2yox/antone/worker/pb"
	"yox2yox/antone/worker/worker"

	"google.golang.org/grpc"
)

var (
	serverAddr = "127.0.0.1:10000"
	port       = "10000"
	testUserId = "client0"
	listener   net.Listener
	grpcServer grpc.Server
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

func UpServer() (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return nil, nil, err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return grpcServer, lis, nil
}

//ValidatebelCodeの生成に成功するか
func TestWorkerEndpoint_GetValidatableCode_Success(t *testing.T) {
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool, false))
		grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()
	defer listen.Close()

	err = datapool.CreateDataPool(testUserId, 0)
	if err != nil {
		t.Fatalf("failed to create datapool %#v", err)
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	orderResult, err := client.GetValidatableCode(ctx, &pb.ValidatableCodeRequest{Bridgeid: "0", Datapoolid: testUserId, Add: 10})
	t.Logf("%#v", orderResult)
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if orderResult == nil {
		t.Fatalf("failed test orderResult is nil")
	}
}

//データプールが存在しないユーザーに対するValidatebleCodeの作成にエラーを出すか
func TestWorkerEndpoint_GetValidatableCodeNotExist_Fail(t *testing.T) {
	//サーバー部
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool, false))
		err = grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()
	defer listen.Close()

	//クライアント部
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.GetValidatableCode(ctx, &pb.ValidatableCodeRequest{Bridgeid: "0", Datapoolid: testUserId, Add: 10})
	if err == nil {
		t.Fatalf("failed to get error %#v", err)
	}
}
