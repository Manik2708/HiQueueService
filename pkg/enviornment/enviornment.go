package enviornment

import "github.com/spf13/viper"

const (
	RABBITMQ_STRING = "RABBITMQ_STRING"
	GRPC_PORT       = "GRPC_PORT"
)

func SetEnviornment() error {
	viper.SetConfigFile(".env")
	return viper.ReadInConfig()
}
