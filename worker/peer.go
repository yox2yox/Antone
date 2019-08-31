package worker

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"
	bpb "yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"
	config "yox2yox/antone/worker/config"
	"yox2yox/antone/worker/datapool"
	pb "yox2yox/antone/worker/pb"
	"yox2yox/antone/worker/worker"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Peer struct {
	WorkerConfig *config.WorkerConfig
	Listener     *net.Listener
	GrpcServer   *grpc.Server
	DataPool     *datapool.Service
}

func New(config *config.WorkerConfig, debug bool) (*Peer, error) {
	peer := &Peer{}

	{ //setup config
		peer.WorkerConfig = config
	}

	{ //setup Server
		lis, err := net.Listen("tcp", peer.WorkerConfig.Server.Addr)
		if err != nil {
			return nil, err
		}
		var opts []grpc.ServerOption
		peer.Listener = &lis
		peer.GrpcServer = grpc.NewServer(opts...)
	}

	{ //setup datapool
		peer.DataPool = datapool.NewService()
	}

	{ //setup worker
		pb.RegisterWorkerServer(peer.GrpcServer, worker.NewEndpoint(peer.DataPool))
	}

	return peer, nil
}

func (p *Peer) Run(ctx context.Context) error {
	//Workerのサインアップ
	if p.WorkerConfig.Bridge.AccountId == "" {
		connBridge, err := grpc.Dial(p.WorkerConfig.Bridge.Addr, grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer connBridge.Close()
		clientAccount := bpb.NewAccountingClient(connBridge)
		ctxAccount, accountCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer accountCancel()
		max := new(big.Int)
		max.SetInt64(int64(p.WorkerConfig.Bridge.AccountRandMax))
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return err
		}
		accountid := fmt.Sprint(n.Int64())
		reqSignup := &bpb.SignupWorkerRequest{Id: accountid, Addr: p.WorkerConfig.Server.Addr}
		_, err = clientAccount.SignupWorker(ctxAccount, reqSignup)
		if err != nil {
			log2.Err.Printf("failed to signup worker\n%#v\n\tworker/peer.go", err)
			return err
		}
		log2.Debug.Printf("an account has been created (id:%s)", accountid)
	}

	//各種サービス起動
	group, ctx := errgroup.WithContext(ctx)

	//サーバー待ち受け開始
	group.Go(func() error {
		err := p.GrpcServer.Serve(*p.Listener)
		if err == context.Canceled || err == grpc.ErrServerStopped || err == http.ErrServerClosed {
			return nil
		}
		return err
	})

	return group.Wait()
}

func (p *Peer) Close() error {
	if p.GrpcServer != nil {
		p.GrpcServer.Stop()
	}
	if p.Listener != nil {
		err := (*p.Listener).Close()
		return err
	}
	return nil
}
