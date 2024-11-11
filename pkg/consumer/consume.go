package consumer

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelResponse[T any] struct {
	Msg *T
	Err error
}

type RecievingRequest struct {
	id   string
	chnl chan ChannelResponse[string]
}

func NewRecievingRequest(id string, chnl chan ChannelResponse[string]) *RecievingRequest {
	rrq := &RecievingRequest{}
	rrq.id = id
	rrq.chnl = chnl
	return rrq
}

func sendErrorToChannel[T any](err error, chnl chan ChannelResponse[T]) {
	c := ChannelResponse[T]{}
	c.Msg = nil
	c.Err = err
	chnl <- c
	close(chnl)
}

type SendingRequest struct {
	id   string
	msg  string
	chnl chan ChannelResponse[bool]
}

func NewSendingRequest(id string, msg string, chnl chan ChannelResponse[bool]) *SendingRequest {
	srq := &SendingRequest{}
	srq.id = id
	srq.chnl = chnl
	srq.msg = msg
	return srq
}

type HiConsumer struct {
	conn     *amqp.Connection
	RecChan  chan RecievingRequest
	SendChan chan SendingRequest
}

func (c *HiConsumer) StartConsuming() error {
	chnl, err := c.conn.Channel()
	if err != nil {
		return err
	}
	for {
		select {
		case rq := <-c.RecChan:
			c.handleRecieve(rq, chnl)
		case rq := <-c.SendChan:
			c.handlePush(rq, chnl)
		}
	}
}

func (c *HiConsumer) handleRecieve(rq RecievingRequest, chnl *amqp.Channel) {
	err := chnl.ExchangeDeclare(rq.id, "fanout", true, false, false, false, nil)
	if err != nil {
		sendErrorToChannel(err, rq.chnl)
		return
	}
	q, err := chnl.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		sendErrorToChannel(err, rq.chnl)
		return
	}
	err = chnl.QueueBind(
		q.Name, // queue name
		"",     // routing key
		rq.id,  // exchange
		false,
		nil)
	if err != nil {
		sendErrorToChannel(err, rq.chnl)
		return
	}
	msgs, err := chnl.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		sendErrorToChannel(err, rq.chnl)
		return
	}
	go func() {
		for msg := range msgs {
			str := string(msg.Body)
			rq.chnl <- ChannelResponse[string]{&str, nil}
		}
		close(rq.chnl)
		_, err = chnl.QueueDelete(q.Name, true, true, false)
		// Push Error to some logging service
	}()
}

func (c *HiConsumer) handlePush(rq SendingRequest, chnl *amqp.Channel) {
	defer close(rq.chnl)
	err := chnl.ExchangeDeclare(rq.id, "fanout", true, false, false, false, nil)
	if err != nil {
		sendErrorToChannel(err, rq.chnl)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	chnl.PublishWithContext(ctx, rq.id, "", true, true, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(rq.msg),
	})
	rsp := true
	rq.chnl <- ChannelResponse[bool]{&rsp, nil}
}
