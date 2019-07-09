package worker

import (
	"context"
	"sync"
	"testing"
	"time"
	config "yox2yox/antone/worker/config"
	pb "yox2yox/antone/worker/pb"

	"google.golang.org/grpc"
)

//---------------------------------------------------------------
//ユーティリティ

func PeerRun(peer *Peer, ctx context.Context) error {

	err := peer.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func InitPeer() (*Peer, error) {
	config, err := config.ReadWorkerConfig()
	if err != nil {
		return nil, err
	}
	peer, err := New(config.Server, true)
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
	if peer.ServerConfig == nil {
		t.Fatalf("failed init peer ServerConfig is nil")
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

//
func TestWorkerEndpoint_ResponseValidatableCode(t *testing.T) {
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

	conn, err := grpc.Dial(peer.ServerConfig.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()

	vCode, err := client.GetValidatableCode(ctxClient, &pb.ValidatableCodeRequest{Bridgeid: "bridge0", Userid: "client0", Add: 10})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if vCode == nil {
		t.Fatalf("failed test validatablecode is nil")
	}

}

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

	conn, err := grpc.Dial(peer.ServerConfig.Addr, grpc.WithInsecure())
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
	if vRes.Db != 11 {
		t.Fatalf("validation failed expected: %d, result:%d", 11, vRes.Db)
	}

}

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

	conn, err := grpc.Dial(peer.ServerConfig.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewWorkerClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()

	dbUpdate, err := client.UpdateDatabase(ctxClient, &pb.DatabaseUpdate{Userid: "client0", Db: 6})
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if dbUpdate == nil {
		t.Fatalf("response database is nil")
	}

}
