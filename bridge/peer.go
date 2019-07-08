package bridge

import (
	"context"
	"errors"
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
	Listner      *net.Listener
	Accounting   *accounting.Service
	Orders       *orders.Service
}

func New(debug bool) (*Peer, error) {

	peer := &Peer{}

	{ //setup config
		config, err := config.ReadBridgeConfig()
		if err != nil {
			return nil, err
		}
		if config == nil {
			return nil, errors.New("config is nil")
		}
		peer.ServerConfig = config.Server
	}

	{ //setup Server
		print("%s", peer.ServerConfig.Addr)
		print("%v", peer.ServerConfig)
		lis, err := net.Listen("tcp", peer.ServerConfig.Addr)
		if err != nil {
			return nil, err
		}
		var opts []grpc.ServerOption
		peer.Listner = &lis
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
		err := p.GrpcServer.Serve(*p.Listner)
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
	if p.Listner != nil {
		err := (*p.Listner).Close()
		return err
	}
	return nil
}
