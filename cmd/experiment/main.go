package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
	"yox2yox/antone/bridge"
	"yox2yox/antone/bridge/config"
	bpb "yox2yox/antone/bridge/pb"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/worker"
	wconfig "yox2yox/antone/worker/config"

	"google.golang.org/grpc"
)

var (
	addrOpt                  = flag.String("a", "", "help message for \"a\" option")
	attackModeOpt            = flag.Int("attack", 0, "help message for \"attack\" option")
	validatorsNumOpt         = flag.Int("validators", 1, "help message for \"validators\" option")
	portOpt                  = flag.Int("port", -1, "help message for \"port\" option")
	numOpt                   = flag.Int("num", 1, "help message for \"port\" option")
	badnumOpt                = flag.Int("bad", 0, "help message for \"bad\" option")
	requestsCountOpt         = flag.Int("req", 1, "help message for \"port\" option")
	faultyFrationOpt         = flag.Float64("fault", 0.4, "help message for \"fault\" option")
	credibilityOpt           = flag.Float64("cred", 0.9, "help message for \"cred\" option")
	resetRateOpt             = flag.Float64("reset", 0, "help message for \"reset\" option")
	watcherOpt               = flag.Bool("watcher", false, "help message for \"watcher\" option")
	blackListingOpt          = flag.Bool("blacklist", false, "help message for \"blacklist\" option")
	stepVotingOpt            = flag.Bool("step", false, "help message for \"step\" option")
	skipValidationOpt        = flag.Bool("skip", false, "help message for \"skip\" option")
	initialReputationOpt     = flag.Int("inirep", 0, "help message for \"inirep\" option")
	sabotagableReputationOpt = flag.Int("sabrep", 0, "help message for \"sabrep\" option")
)

func main() {

	flag.Parse()

	defer log2.Close()

	log2.Debug.Println("starting main function...")

	errch := make(chan struct{})
	var wg sync.WaitGroup
	var wgBridge sync.WaitGroup

	config, err := config.ReadBridgeConfig()
	if err != nil {
		fmt.Printf("FATAL %s [] Failed to read config", time.Now())
	}

	if *numOpt <= 0 {
		return
	}

	baseport := -1
	if *portOpt > 0 {
		baseport = *portOpt + 1
	}
	badnum := 0
	if *badnumOpt > 0 {
		badnum = *badnumOpt
	}

	if *addrOpt != "" {
		config.Server.Addr = *addrOpt + ":" + strconv.Itoa(*portOpt)
	}

	if *validatorsNumOpt != 1 {
		config.Order.NeedValidationNum = *validatorsNumOpt
	}

	requestsCount := 1
	if *requestsCountOpt != 1 {
		requestsCount = *requestsCountOpt
	}

	//ブリッジ起動
	peerBridge, err := bridge.New(config, false, *faultyFrationOpt, *credibilityOpt, *resetRateOpt, *watcherOpt, *blackListingOpt, *stepVotingOpt, *skipValidationOpt, *initialReputationOpt, *attackModeOpt, *sabotagableReputationOpt)
	if err != nil {
		fmt.Printf("FATAL %s [] Failed to initialize peer", time.Now())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		wgBridge.Add(1)
		err = peerBridge.Run(ctx)
		if err != nil {
			fmt.Printf("FATAL %s [] Peer was killed %#v", time.Now(), err)
		}
		wgBridge.Done()
	}()

	//ワーカー起動
	wPeers := []*worker.Peer{}
	for i := 0; i < *numOpt; i++ {
		workerConfig, err := wconfig.ReadWorkerConfig()
		if err != nil {
			log2.Err.Println("failed to read config")
			return
		}

		if *addrOpt != "" {
			addr := *addrOpt
			if baseport > 0 {
				port := strconv.Itoa(baseport)
				addr += ":" + port
				baseport += 1
			}
			workerConfig.Server.Addr = addr
		}

		badmode := false
		if badnum > 0 {
			badmode = true
			badnum--
		}
		log2.Debug.Printf("Addr:%s Bridge:%s Id:%s BadMode:%v", workerConfig.Server.Addr, workerConfig.Bridge.Addr, workerConfig.Bridge.AccountId, badmode)
		peer, err := worker.New(workerConfig, false, badmode, *attackModeOpt)
		if err != nil {
			log2.Err.Printf("failed to initialize peer %#v", err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		wPeers = append(wPeers, peer)
		go func() {
			wg.Add(1)
			err = peer.Run(ctx)
			if err != nil {
				log2.Err.Printf("peer was killed \n%#v\n", err)
			}
			wg.Done()
		}()
	}

	//クライアント起動

	myUserId := "client3"
	myAddr := "localhost"
	holdersnum := 3

	connBridge, err := grpc.Dial(config.Server.Addr, grpc.WithInsecure())
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
		for i := 0; i < requestsCount; i++ {
			ctxOrder, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			waitForValidation := false
			if *skipValidationOpt == false || i == requestsCount-1 {
				waitForValidation = true
			}
			req := &bpb.ValidatableCodeRequest{Datapoolid: myDatapool.Datapoolid, Add: 1, WaitForValidation: waitForValidation}
			clientOrder.RequestValidatableCode(ctxOrder, req)
			log2.Debug.Printf("Got Validatablecode in client")
		}
		close(errch)
	}()

	//コマンドをスキャン
	scanner := bufio.NewScanner(os.Stdin)
	stoped := false
	go func() {
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
	}()

	select {
	case <-errch:
		log2.Debug.Println("clossing peer...")
		peerBridge.Close()
		wgBridge.Wait()
		for _, p := range wPeers {
			p.Close()
		}
		wg.Wait()
		return
	}
}
