package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"yox2yox/antone/internal/log2"
	"yox2yox/antone/worker"
	"yox2yox/antone/worker/config"
)

var (
	addrOpt       = flag.String("a", "", "help message for \"a\" option")
	bridgeAddrOpt = flag.String("bridge", "", "help message for \"bridge\" option")
	idOpt         = flag.String("id", "", "help message for \"id\" option")
	badModeOpt    = flag.Bool("bad", false, "help message for \"bad\" option")
)

func main() {

	flag.Parse()
	badmode := false

	defer log2.Close()

	errch := make(chan struct{})

	log2.Debug.Println("start main function")
	config, err := config.ReadWorkerConfig()
	if *addrOpt != "" {
		config.Server.Addr = *addrOpt
	}
	if *bridgeAddrOpt != "" {
		config.Bridge.Addr = *bridgeAddrOpt
	}
	if *idOpt != "" {
		config.Bridge.AccountId = *idOpt
	}
	if *badModeOpt == true {
		badmode = true
	}

	if err != nil {
		log2.Err.Println("failed to read config")
		return
	}
	peer, err := worker.New(config, false, badmode)
	if err != nil {
		log2.Err.Println("failed to initialize peer")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err = peer.Run(ctx)
		if err != nil {
			log2.Err.Printf("peer was killed \n%#v\n", err)
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
		peer.Close()
		return
	}

}
