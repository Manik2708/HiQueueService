package main

import (
	"fmt"
	"os"

	"github.com/Manik2708/HiQueueService/pkg/consumer"
	"github.com/Manik2708/HiQueueService/pkg/enviornment"
	"github.com/Manik2708/HiQueueService/pkg/server"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

func main() {
	err := enviornment.SetEnviornment()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	c := &consumer.HiConsumer{}
	conn, err := amqp.Dial(viper.GetString(enviornment.RABBITMQ_STRING))
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	c.Conn = conn
	c.SendChan = make(chan consumer.SendingRequest)
	c.RecChan = make(chan consumer.RecievingRequest)
	go c.StartConsuming()
	s := server.CoreServer{}
	err = s.New(viper.GetString(enviornment.GRPC_PORT), c)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
}
