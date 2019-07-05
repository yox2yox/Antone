package order

import (
	"context"
	"errors"
	"math/rand"
	"time"
	pb "yox2yox/antone/bridge/pb"
)

type Worker struct {
	addr       string
	reputation int
}

type Endpoint struct {
	workers []Worker
	pickNum int
	holder  map[string][]int
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		workers: []Worker{
			Worker{
				ip:         "localhost",
				reputation: 0,
			},
		},
		pickNum: 3,
		holder: map[string][]int{
			"0": []int{
				0,
			},
		},
	}
}

func (e *Endpoint) CreateOrder(ctx context.Context, workRequest *pb.WorkRequest) (*pb.OrderInfo, error) {
	if len(e.workers) <= e.pickNum {
		return nil, errors.New("There is not enough Wokers")
	}
	rand.Seed(time.Now().UnixNano())

	holderid := e.holder[workRequest.Userid][0]

	picked := []Worker{}
	for i := 0; i < e.pickNum; i++ {
		picked = append(picked, e.workers[rand.Intn(len(e.workers))])
	}

	//Send OrderRequest

	return &pb.OrderInfo{
		Addr:               e.workers[holderid].addr,
		DatabaseMarkleRoot: "",
		ScriptMarkleRoot:   "",
	}, nil
}
