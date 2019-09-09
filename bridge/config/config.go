package config

type ServerConfig struct {
	Addr string
	Port string
}

type OrderConfig struct {
	ErrorProbability  float64
	NeedValidationNum int //承認されるまでに必要なバリデーターの数
}

type BridgeConfig struct {
	Server *ServerConfig
	Order  *OrderConfig
}

func ReadBridgeConfig() (*BridgeConfig, error) {

	config := new(BridgeConfig)

	config.Server = &ServerConfig{}

	config.Server.Addr = "localhost:10000"
	config.Server.Port = "10000"

	config.Order = &OrderConfig{}

	config.Order.ErrorProbability = 0.1
	config.Order.NeedValidationNum = 1

	return config, nil
}
