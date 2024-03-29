package bridge

import (
	"context"
	"sync"
	"testing"
	"time"
	"os"
	"yox2yox/antone/internal/log2"
	config "yox2yox/antone/bridge/config"
	pb "yox2yox/antone/bridge/pb"

	"google.golang.org/grpc"
)

var (
	testClientId   = "client0"
	testWorkerId   = "worker0"
	testWorkerAddr = "localhost:10000"
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

func PeerRun(peer *Peer, ctx context.Context) error {

	err := peer.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func InitPeer() (*Peer, error) {
	config, err := config.ReadBridgeConfig()
	if err != nil {
		return nil, err
	}
	peer, err := New(config, true)
	if err != nil {
		return nil, err
	}
	return peer, nil
}

func TestPeer_Initialize(t *testing.T) {
	peer, err := InitPeer()
	defer peer.Close()
	if err != nil {
		t.Fatalf("failed init peer %#v", err)
	}
	if peer.GrpcServer == nil {
		t.Fatalf("failed init peer GrpcServer is nil")
	}
	if peer.Config.Server == nil {
		t.Fatalf("failed init peer ServerConfig is nil")
	}
	if peer.Listener == nil {
		t.Fatalf("failed init peer Listener is nil")
	}
	if peer.Accounting == nil {
		t.Fatalf("failed init peer Accounting is nil")
	}
	if peer.Orders == nil {
		t.Fatalf("failed init peer Accounting is nil")
	}
}

func TestPeer_RunAndClose(t *testing.T) {
	wg := sync.WaitGroup{}

	peer, err := InitPeer()
	if err != nil {
		t.Fatalf("failed to initialize peer %#v", err)
	}

	wg.Add(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err := PeerRun(peer, ctx)
		if err != nil {
			t.Fatalf("failed to close peer %#v", err)
		}
		defer wg.Done()
	}()
	cancel()
	err = peer.Close()
	if err != nil {
		t.Fatalf("failed to close peer %#v", err)
	}
	wg.Wait()
}

func TestOrderEndpoint_ResponseValidatableCode(t *testing.T) {

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

	_, err = peer.Accounting.CreateNewWorker(testWorkerId, testWorkerAddr)
	if err != nil {
		t.Fatalf("failed to create worker %#v", err)
	}

	createdDp, err := peer.Datapool.CreateDatapool(testClientId, 1)
	if err != nil {
		t.Fatalf("want no error,but error %#v", err)
	}

	conn, err := grpc.Dial(peer.Config.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewOrdersClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()
	vCode, err := client.RequestValidatableCode(ctxClient, &pb.ValidatableCodeRequest{Datapoolid: createdDp.DatapoolId, Add: 10})

	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if vCode == nil {
		t.Fatalf("failed test orderinfo is nil")
	}

}
