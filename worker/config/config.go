package config

type ServerConfig struct {
	Addr string
	Port string
}

type WorkerConfig struct {
	Server *ServerConfig
}

func ReadWorkerConfig() (*WorkerConfig, error) {

	config := new(WorkerConfig)
	config.Server = &ServerConfig{}

	config.Server.Addr = "127.0.0.1:10001"
	config.Server.Port = "10001"

	return config, nil
}
