package bridge

import (
	"context"
	"sync"
	"testing"
	"time"
	config "yox2yox/antone/bridge/config"
	pb "yox2yox/antone/bridge/pb"

	"google.golang.org/grpc"
)

func TestPeerInit(t *testing.T) {
	peer, err := New(true)
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
	if peer.Listner == nil {
		t.Fatalf("failed init peer Listener is nil")
	}
	if peer.Accounting == nil {
		t.Fatalf("failed init peer Accounting is nil")
	}
	if peer.Orders == nil {
		t.Fatalf("failed init peer Accounting is nil")
	}
}

func TestPeerRunAndClose(t *testing.T) {
	wg := sync.WaitGroup{}
	peer, err := New(true)
	if err != nil {
		t.Fatalf("failed peer init %#v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = peer.Run(ctx)
		if err != nil {
			t.Fatalf("failed peer run %#v", err)
		}
	}()
	cancel()
	err = peer.Close()
	if err != nil {
		t.Fatalf("failed close peer %#v", err)
	}
	wg.Wait()
}

func TestPeerRunOrderEndpoint(t *testing.T) {
	peer, err := New(true)
	if err != nil {
		t.Fatalf("failed peer init %#v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		err = peer.Run(ctx)
		if err != nil {
			t.Fatalf("failed peer run %#v", err)
		}
	}()
	defer cancel()
	defer peer.Close()
	config, err := config.ReadBridgeConfig()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	conn, err := grpc.Dial(config.Server.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	defer conn.Close()
	client := pb.NewOrdersClient(conn)
	ctxClient, cancelClient := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelClient()
	vCode, err := client.RequestValidatableCode(ctxClient, &pb.ValidatableCodeRequest{Userid: "0", Add: 10})

	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if vCode == nil {
		t.Fatalf("failed test orderinfo is nil")
	}

}
