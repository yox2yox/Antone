package worker_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
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

func UpServer() (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		return nil, nil, err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return grpcServer, lis, nil
}

//データプールの作成に成功するか
func TestWorkerEndpoint_CreateNewDataPool_Success(t *testing.T) {
	//サーバー部
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool))
		err = grpcServer.Serve(listen)
		if err != nil && err != context.Canceled && err != grpc.ErrServerStopped && err != http.ErrServerClosed {
			t.Fatalf("failed test %#v", err.Error())
		}
	}()
	defer grpcServer.Stop()

	//クライアント部
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	createResult, err := client.CreateNewDatapool(ctx, &pb.DatapoolInfo{Userid: testUserId})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if createResult == nil {
		t.Fatalf("failed test updateResult is nil")
	}
}

//既に存在するデータプールの作成に正しいエラーを出すか
func TestWorkerEndpoint_CreateExistDataPool_Fail(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	//サーバー部
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool))
		err = grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
		defer wg.Done()
	}()

	//データプール作成
	err = datapool.CreateNewDataPool(testUserId)
	if err != nil {
		t.Fatalf("failed to create datapool %#v", err)
	}

	//クライアント部
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	_, err = client.CreateNewDatapool(ctx, &pb.DatapoolInfo{Userid: testUserId})
	if err == nil {
		t.Fatalf("failed to get error %#v", err)
	}

	conn.Close()
	cancel()
	grpcServer.Stop()
	listen.Close()
	wg.Wait()
}

//ValidatebelCodeの生成に成功するか
func TestWorkerEndpoint_GetValidatableCode_Success(t *testing.T) {
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool))
		grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()
	defer listen.Close()

	err = datapool.CreateNewDataPool(testUserId)
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
	orderResult, err := client.GetValidatableCode(ctx, &pb.ValidatableCodeRequest{Bridgeid: "0", Userid: testUserId, Add: 10})
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
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool))
		err = grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()

	//クライアント部
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.GetValidatableCode(ctx, &pb.ValidatableCodeRequest{Bridgeid: "0", Userid: testUserId, Add: 10})
	if err == nil {
		t.Fatalf("failed to get error %#v", err)
	}
}

//データプールの更新に成功するか
func TestWorkerEndpoint_UpdateDatabase_Success(t *testing.T) {
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
		return
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool))
		err = grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err.Error())
		}
	}()
	defer grpcServer.Stop()

	err = datapool.CreateNewDataPool(testUserId)
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
	updateResult, err := client.UpdateDatapool(ctx, &pb.DatapoolUpdate{Userid: testUserId, Pool: 50})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if updateResult == nil {
		t.Fatalf("failed test updateResult is nil")
	}
}

//存在しないデータプールの更新に対してエラーを出すか
func TestWorkerEndpoint_UpdateDataPoolNotExist_Fail(t *testing.T) {
	//サーバー部
	datapool := datapool.NewService()
	grpcServer, listen, err := UpServer()
	if err != nil {
		t.Fatalf("failed to up server %#v", err)
	}
	go func() {
		pb.RegisterWorkerServer(grpcServer, worker.NewEndpoint(datapool))
		err = grpcServer.Serve(listen)
		if err != nil {
			t.Fatalf("failed test %#v", err)
		}
	}()
	defer grpcServer.Stop()

	//クライアント部
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.UpdateDatapool(ctx, &pb.DatapoolUpdate{Userid: testUserId, Pool: 50})
	if err == nil {
		t.Fatalf("failed to get error %#v", err)
	}

}
