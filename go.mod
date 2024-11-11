module github.com/Manik2708/HiQueueService

go 1.23.0

require github.com/rabbitmq/amqp091-go v1.10.0 // direct

require (
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.1
)

require (
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241104194629-dd2ea8efbc28 // indirect
)

replace(
	github.com/Manik2708/HiQueueService/pkg => ./pkg
)
