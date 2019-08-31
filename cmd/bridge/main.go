package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
	"yox2yox/antone/bridge"
	"yox2yox/antone/bridge/config"
)

func main() {

	config, err := config.ReadBridgeConfig()
	if err != nil {
		fmt.Printf("FATAL %s [] Failed to read config", time.Now())
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
		fmt.Printf("Antone > ")
	}
	peer.Close()

}
