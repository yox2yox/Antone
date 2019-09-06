package worker_test

import (
	"context"
	"sync"
	"testing"
	"time"
	"yox2yox/antone/worker"
	config "yox2yox/antone/worker/config"
	pb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

var (
	testUserId = "client0"
)

//---------------------------------------------------------------
//共通部

func PeerRun(peer *worker.Peer, ctx context.Context) error {
	err := peer.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func InitPeer() (*worker.Peer, error) {
	config, err := config.ReadWorkerConfig()
	if err != nil {
		return nil, err
	}
	peer, err := worker.New(config, true)
	if err != nil {
		return nil, err
	}
	return peer, nil
}

//---------------------------------------------------------------
//テストコード

//ピア初期化テスト
func TestPeerInit(t *testing.T) {
	peer, err := InitPeer()
	defer peer.Close()
	if err != nil {
		t.Fatalf("failed init peer %#v", err)
	}
	if peer.GrpcServer == nil {
		t.Fatalf("failed init peer GrpcServer is nil")
	}
	if peer.WorkerConfig == nil {
		t.Fatalf("failed init peer WorkerConfig is nil")
	}
	if peer.Listener == nil {
		t.Fatalf("failed init peer Listener is nil")
	}
}

//ピアの起動と停止テスト
func TestPeer_RunAndClose(t *testing.T) {
	wg := sync.WaitGroup{}

	peer, err := InitPeer()
	if err != nil {
		t.Fatalf("failed initializing peer %#v", err)
	}

	wg.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err := PeerRun(peer, ctx)
		if err != nil {
			t.Fatalf("failed closing peer %#v", err)
		}
		defer wg.Done()
	}()
	cancel()
	err = peer.Close()
	if err != nil {
		t.Fatalf("failed close peer %#v", err)
	}
	wg.Wait()
}

//GetValidatableCodeリクエストを正しく処理するか
func TestWorkerEndpoint_ResponseValidatableCode(t *testing.T) {
	peer, err := InitPeer()
	if err != nil {
		t.Fatalf("failed to initialize peer %#v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err := PeerRun(peer, ctx)
		if err != nil {
			t.Fatalf("failed to close peer %#v", err)
		}
	}()
	defer cancel()
	defer peer.Close()

	conn, err := grpc.Dial(peer.WorkerConfig.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to get connection %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()

	_, err = client.GetValidatableCode(ctxClient, &pb.ValidatableCodeRequest{Bridgeid: "bridge0", Userid: testUserId, Add: 10})
	if err == nil {
		t.Fatalf("want error but nil")
	}

	peer.DataPool.CreateDataPool(testUserId, 0)

	vCode, err := client.GetValidatableCode(ctxClient, &pb.ValidatableCodeRequest{Bridgeid: "bridge0", Userid: testUserId, Add: 10})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if vCode == nil {
		t.Fatalf("failed test validatablecode is nil")
	}

}

//OrderValidationを正しく処理するか
func TestWorkerEndpoint_Validate(t *testing.T) {
	peer, err := InitPeer()
	if err != nil {
		t.Fatalf("failed initializing peer %#v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err := PeerRun(peer, ctx)
		if err != nil {
			t.Fatalf("failed closing peer %#v", err)
		}
	}()
	defer cancel()
	defer peer.Close()

	conn, err := grpc.Dial(peer.WorkerConfig.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()

	vRes, err := client.OrderValidation(ctxClient, &pb.ValidatableCode{Data: 1, Add: 10})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if vRes == nil {
		t.Fatalf("failed test validation response is nil")
	}
	if vRes.Pool != 11 {
		t.Fatalf("validation failed expected: %d, result:%d", 11, vRes.Pool)
	}

}

//UpdateDatabaseを正しく処理するか
func TestWorkerEndpoint_UpdateDb(t *testing.T) {
	peer, err := InitPeer()
	if err != nil {
		t.Fatalf("failed initializing peer %#v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err := PeerRun(peer, ctx)
		if err != nil {
			t.Fatalf("failed closing peer %#v", err)
		}
	}()
	defer cancel()
	defer peer.Close()

	conn, err := grpc.Dial(peer.WorkerConfig.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewDatapoolClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()

	dbUpdate, err := client.UpdateDatapool(ctxClient, &pb.DatapoolContent{Id: testUserId, Data: 6})
	if err == nil {
		t.Fatalf("want error, but nil")
	}

	peer.DataPool.CreateDataPool(testUserId, 0)

	dbUpdate, err = client.UpdateDatapool(ctxClient, &pb.DatapoolContent{Id: testUserId, Data: 6})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if dbUpdate == nil {
		t.Fatalf("response database is nil")
	}

}
