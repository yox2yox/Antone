package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"
	bpb "yox2yox/antone/bridge/pb"

	"google.golang.org/grpc"
)

func main() {

	bridgeAddr := ""
	myUserId := "client0"
	myAddr := "192.168.25.10"
	var sendIntervalMilliSec time.Duration = 500
	holdersnum := 3
	doneChannel := make(chan struct{})

	flag.Parse()
	args := flag.Args()

	if len(args) > 1 && args[0] != "" {
		bridgeAddr = args[0]
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

	go func() {
		for {
			select {
			case <-doneChannel:
				return
			default:
				ctxOrder, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				req := &bpb.ValidatableCodeRequest{Datapoolid: myDatapool.Datapoolid, Add: 1}
				clientOrder.RequestValidatableCode(ctxOrder, req)
			}
			time.Sleep(sendIntervalMilliSec * time.Millisecond)
		}
	}()

	//コマンドをスキャン
	scanner := bufio.NewScanner(os.Stdin)
	stoped := false
	fmt.Printf("Antone > ")
	for stoped == false && scanner.Scan() {
		cmd := scanner.Text()
		switch cmd {
		case "stop":
			stoped = true
			break
		default:
			fmt.Printf("%s: command not found\n", cmd)
		}
		if stoped == false {
			fmt.Printf("Antone > ")
		}
	}
	close(doneChannel)

}
