package bridge

import (
	"context"
	"net"
	"net/http"
	"yox2yox/antone/bridge/accounting"
	config "yox2yox/antone/bridge/config"
	"yox2yox/antone/bridge/datapool"
	"yox2yox/antone/bridge/orders"
	pb "yox2yox/antone/bridge/pb"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"
)

type Peer struct {
	GrpcServer *grpc.Server
	Config     *config.BridgeConfig
	Addr       string
	port       string
	Listener   *net.Listener
	Accounting *accounting.Service
	Datapool   *datapool.Service
	Orders     *orders.Service
}

func New(config *config.BridgeConfig, debug bool, faultyFraction float64, credibilityThreshold float64, reputationResetRate float64, setWathcer bool, blackListing bool) (*Peer, error) {

	peer := &Peer{}

	{ //setup config
		peer.Config = config
	}

	{ //setup Server
		lis, err := net.Listen("tcp", peer.Config.Server.Addr)
		if err != nil {
			return nil, err
		}
		var opts []grpc.ServerOption
		peer.Listener = &lis
		peer.GrpcServer = grpc.NewServer(opts...)
	}

	{ //setup Accounting
		peer.Accounting = accounting.NewService(debug, faultyFraction, credibilityThreshold, reputationResetRate, blackListing)
		pb.RegisterAccountingServer(peer.GrpcServer, accounting.NewEndpoint(peer.Accounting))
	}

	{ //setup Datapool
		peer.Datapool = datapool.NewService(peer.Accounting, debug)
		pb.RegisterDatapoolServer(peer.GrpcServer, datapool.NewEndpoint(peer.Datapool, peer.Accounting))
	}

	{ //setup orders
		peer.Orders = orders.NewService(peer.Accounting, peer.Datapool, debug, peer.Config.Order.CalcErrorRate, setWathcer)
		pb.RegisterOrdersServer(peer.GrpcServer, orders.NewEndpoint(peer.Config.Order, peer.Orders, peer.Accounting))
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
	group.Go(func() error {
		p.Orders.Run()
		return nil
	})
	return group.Wait()
}

func (p *Peer) Close() error {
	if p.GrpcServer != nil {
		p.GrpcServer.Stop()
	}
	if p.Orders != nil {
		p.Orders.Stop()
	}
	if p.Listener != nil {
		err := (*p.Listener).Close()
		return err
	}
	return nil
}
