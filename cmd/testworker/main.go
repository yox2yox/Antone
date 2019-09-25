package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/worker"
	"yox2yox/antone/worker/config"
)

var (
	addrOpt       = flag.String("a", "", "help message for \"a\" option")
	portOpt       = flag.Int("port", -1, "help message for \"port\" option")
	numOpt        = flag.Int("num", 1, "help message for \"port\" option")
	badnumOpt     = flag.Int("bad", 0, "help message for \"bad\" option")
	bridgeAddrOpt = flag.String("bridge", "", "help message for \"bridge\" option")
)

func main() {

	flag.Parse()

	defer log2.Close()

	errch := make(chan struct{})

	log2.Debug.Println("start main function")

	if *numOpt <= 0 {
		return
	}

	baseport := -1
	if *portOpt > 0 {
		baseport = *portOpt
	}
	badnum := 0
	if *badnumOpt > 0 {
		badnum = *badnumOpt
	}
	for i := 0; i < *numOpt; i++ {
		workerConfig, err := config.ReadWorkerConfig()
		if err != nil {
			log2.Err.Println("failed to read config")
			return
		}
		if *bridgeAddrOpt != "" {
			workerConfig.Bridge.Addr = *bridgeAddrOpt
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
		peer, err := worker.New(workerConfig, false, badmode)
		if err != nil {
			log2.Err.Println("failed to initialize peer")
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		defer peer.Close()
		go func() {
			err = peer.Run(ctx)
			if err != nil {
				log2.Err.Printf("peer was killed \n%#v\n", err)
			}
			close(errch)
		}()
	}

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
				close(errch)
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
		return
	}

}
