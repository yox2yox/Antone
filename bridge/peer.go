package bridge

import (
	"context"
	"net"
	"net/http"
	"yox2yox/antone/bridge/accounting"
	config "yox2yox/antone/bridge/config"
	"yox2yox/antone/bridge/orders"
	pb "yox2yox/antone/bridge/pb"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"
)

type Peer struct {
	GrpcServer   *grpc.Server
	ServerConfig *config.ServerConfig
	Addr         string
	port         string
	Listener     *net.Listener
	Accounting   *accounting.Service
	Orders       *orders.Service
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

	{ //setup Accounting
		peer.Accounting = accounting.NewService()
	}

	{ //setup orders
		peer.Orders = orders.NewService(peer.Accounting, debug)
		pb.RegisterOrdersServer(peer.GrpcServer, orders.NewEndpoint(peer.Accounting))
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
