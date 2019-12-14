package worker

import (
	"context"
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
	WorkerConfig                *config.WorkerConfig
	Listener                    *net.Listener
	GrpcServer                  *grpc.Server
	DataPool                    *datapool.Service
	WithoutConnectRemoteForTest bool
	BadMode                     bool
}

func New(config *config.WorkerConfig, debug bool, badmode bool, attackMode int) (*Peer, error) {
	peer := &Peer{}

	{ //setup debug mode
		peer.WithoutConnectRemoteForTest = debug
		peer.BadMode = badmode
	}

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
		pb.RegisterDatapoolServer(peer.GrpcServer, datapool.NewEndpoint(peer.DataPool))
	}

	{ //setup worker
		pb.RegisterWorkerServer(peer.GrpcServer, worker.NewEndpoint(peer.DataPool, badmode, attackMode))
	}

	return peer, nil
}

func (p *Peer) Run(ctx context.Context) error {
	//Workerのサインアップ
	if p.WithoutConnectRemoteForTest == false {
		if p.WorkerConfig.Bridge.AccountId == "" {
			connBridge, err := grpc.Dial(p.WorkerConfig.Bridge.Addr, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer connBridge.Close()
			clientAccount := bpb.NewAccountingClient(connBridge)
			ctxAccount, accountCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer accountCancel()
			accountid := p.WorkerConfig.Server.Addr
			reqSignup := &bpb.SignupWorkerRequest{Id: accountid, Addr: p.WorkerConfig.Server.Addr, IsBad: p.BadMode}
			_, err = clientAccount.SignupWorker(ctxAccount, reqSignup)
			if err != nil {
				log2.Err.Printf("failed to signup worker\n%#v\n\tworker/peer.go", err)
				return err
			}
			log2.Debug.Printf("an account has been created (id:%s)", accountid)
		}
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
	log2.Debug.Printf("closing worker server... Addr:%s", p.WorkerConfig.Server.Addr)
	if p.GrpcServer != nil {
		p.GrpcServer.Stop()
	}
	if p.Listener != nil {
		err := (*p.Listener).Close()
		return err
	}
	return nil
}
