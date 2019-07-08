package config

type ServerConfig struct {
	Addr string
	Port string
}

type BridgeConfig struct {
	Server ServerConfig
}

func ReadBridgeConfig() (*BridgeConfig, error) {

	config := new(BridgeConfig)
	config.Server = ServerConfig{}

	config.Server.Addr = "127.0.0.1:10000"
	config.Server.Port = "10000"

	return config, nil
}
