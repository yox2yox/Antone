package worker

import (
	"context"
	"net"
	"net/http"
	config "yox2yox/antone/worker/config"
	pb "yox2yox/antone/worker/pb"
	"yox2yox/antone/worker/worker"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Peer struct {
	ServerConfig *config.ServerConfig
	Listener     *net.Listener
	GrpcServer   *grpc.Server
}

func New(config *config.ServerConfig, debug bool) (*Peer, error) {
	peer := &Peer{}

	{ //setup config
		peer.ServerConfig = config
	}

	{ //setup Server
		lis, err := net.Listen("tcp", peer.ServerConfig.Addr)
		if err != nil {
			return nil, err
		}
		var opts []grpc.ServerOption
		peer.Listener = &lis
		peer.GrpcServer = grpc.NewServer(opts...)
	}

	{ //setup worker
		pb.RegisterWorkerServer(peer.GrpcServer, worker.NewEndpoint())
	}

	return peer, nil
}

func (p *Peer) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
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
