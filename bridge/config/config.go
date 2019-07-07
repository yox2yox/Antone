package config

import (
	"encoding/json"
	"io/ioutil"
)

type ServerConfig struct {
	Addr string `json:"addr"`
	Port string `json:"port"`
}

type Config struct {
	Server ServerConfig `json:"server"`
}

var (
	filename = "config.json"
)

func ReadConfig() (*Config, error) {
	config := new(Config)

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
