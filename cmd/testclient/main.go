package main

import (
	"context"
	"flag"
	"time"
	bpb "yox2yox/antone/bridge/pb"

	"google.golang.org/grpc"
)

var (
	bridgeAddrOpt    = flag.String("bridge", "", "help message for \"bridge\" option")
	intervalOpt      = flag.Int("intv", -1, "help message for \"intv\" option")
	requestsCountOpt = flag.Int("req", 1, "help message for \"req\" option")
)

func main() {

	bridgeAddr := ""
	myUserId := "client3"
	myAddr := "192.168.25.10"
	holdersnum := 3

	flag.Parse()

	if *bridgeAddrOpt != "" {
		bridgeAddr = *bridgeAddrOpt
	}

	//ブリッジクライアント作成
	connBridge, err := grpc.Dial(bridgeAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer connBridge.Close()
	clientOrder := bpb.NewOrdersClient(connBridge)
	clientAccounting := bpb.NewAccountingClient(connBridge)
	clientDatapool := bpb.NewDatapoolClient(connBridge)

	ctxSignup, cancelSignup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSignup()
	reqSignup := &bpb.SignupClientRequest{Id: myUserId, Addr: myAddr}
	clientAccounting.SignupClient(ctxSignup, reqSignup)

	ctxCreateDatapool, cancelCreateDatapool := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCreateDatapool()
	reqCreateDatapool := &bpb.CreateRequest{Userid: myUserId, Holdersnum: int32(holdersnum)}
	myDatapool, err := clientDatapool.CreateDatapool(ctxCreateDatapool, reqCreateDatapool)
	if err != nil {
		panic(err)
	}

	for i := 0; i < *requestsCountOpt; i += 1 {
		ctxOrder, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		req := &bpb.ValidatableCodeRequest{Datapoolid: myDatapool.Datapoolid, Add: 1}
		clientOrder.RequestValidatableCode(ctxOrder, req)
	}

}
