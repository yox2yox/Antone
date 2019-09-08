package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"time"
	"yox2yox/antone/bridge"
	"yox2yox/antone/bridge/config"
	"yox2yox/antone/internal/log2"
)

func main() {

	flag.Parse()
	args := flag.Args()

	defer log2.Close()

	config, err := config.ReadBridgeConfig()
	if err != nil {
		fmt.Printf("FATAL %s [] Failed to read config", time.Now())
	}

	if len(args) > 1 && args[0] != "" {
		config.Server.Addr = args[0]
	}

	peer, err := bridge.New(config, false)
	if err != nil {
		fmt.Printf("FATAL %s [] Failed to initialize peer", time.Now())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err = peer.Run(ctx)
		if err != nil {
			fmt.Printf("FATAL %s [] Peer was killed %#v", time.Now(), err)
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
	peer.Close()

}
