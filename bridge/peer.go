package order

import (
	"context"
	"fmt"
	"net"
	"yox2yox/antone/bridge/accounting"
	"yox2yox/antone/bridge/config"
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
		config, err := config.ReadConfig()
		if err != nil {
			return nil, err
		}
		peer.ServerConfig = &config.Server
	}

	{ //setup Server
		lis, err := net.Listen("tcp", fmt.Sprintf("%s", peer.ServerConfig.Addr))
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
		return p.GrpcServer.Serve(*p.Listner)
	})
	return group.Wait()
}
