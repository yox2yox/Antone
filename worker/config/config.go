package config

type ServerConfig struct {
	Addr string
	Port string
}

type BridgeConnectionConfig struct {
	Addr           string
	Port           string
	AccountId      string
	AccountRandMax int
}

type WorkerConfig struct {
	Server *ServerConfig
	Bridge *BridgeConnectionConfig
}

func ReadWorkerConfig() (*WorkerConfig, error) {

	config := new(WorkerConfig)
	config.Server = &ServerConfig{}
	config.Bridge = &BridgeConnectionConfig{}

	config.Server.Addr = "127.0.0.1:10001"
	config.Server.Port = "10001"

	config.Bridge.Addr = "127.0.0.1:10000"
	config.Bridge.Port = "10000"
	config.Bridge.AccountId = ""
	config.Bridge.AccountRandMax = 100000

	return config, nil
}
