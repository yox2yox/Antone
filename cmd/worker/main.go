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

func main() {

	flag.Parse()
	args := flag.Args()
	badmode := false

	errch := make(chan struct{})

	log2.Debug.Println("start main function")
	config, err := config.ReadWorkerConfig()
	if len(args) > 1 && args[0] != "" {
		config.Server.Addr = args[0]
	}
	if len(args) > 2 && args[1] != "" {
		config.Bridge.Addr = args[1]
	}
	if len(args) > 3 {
		config.Bridge.AccountId = args[2]
	}
	if len(args) > 4 && args[3] == "true" {
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
