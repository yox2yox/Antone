package config

import (
	"encoding/json"
	"io/ioutil"
)

type ServerConfig struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
}

type BridgeConfig struct {
	Server ServerConfig `json:"server"`
}

var (
	filename = "bridge_config.json"
)

func ReadBridgeConfig() (*BridgeConfig, error) {
	config := new(BridgeConfig)

	jsonString, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(jsonString, config)
	if err != nil {
		return config, err
	}
	return config, nil
}
