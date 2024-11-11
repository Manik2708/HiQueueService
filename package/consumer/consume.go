package consumer

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RecievingRequest struct{
	id string
	chnl chan string
	errChnl chan error
}

type SendingRequest struct{
	id string
	msg string
	chnl chan bool
	errChnl chan error
}

type HiConsumer struct {
	conn *amqp.Connection
	recChan chan RecievingRequest
	sendChan chan SendingRequest
}

func (c *HiConsumer) StartConsuming () error {
	chnl, err := c.conn.Channel()
	if err != nil{
		return err
	}
	for{
		select{
		case rq := <- c.recChan:
			c.handleRecieve(rq, chnl)
		case rq := <- c.sendChan:
			c.handlePush(rq, chnl)
		}
	}
}

func (c *HiConsumer) handleRecieve(rq RecievingRequest, chnl *amqp.Channel){
	err := chnl.ExchangeDeclare(rq.id, "fanout", true, false, false, false, nil)
			if err != nil{
				rq.errChnl <- err
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
			if err != nil{
				rq.errChnl <- err
				return
			}
			err = chnl.QueueBind(
				q.Name,        // queue name
				"",             // routing key
				rq.id, // exchange
				false,
				nil)
			if err != nil{
				rq.errChnl <- err
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
		if err != nil{
			rq.errChnl <- err
			return
	}
		go func() {
			for msg := range msgs{
				rq.chnl <- string(msg.Body)
			}
			_, err = chnl.QueueDelete(q.Name, true, true, false)
			// Push Error to some logging service
		}()
}

func (c *HiConsumer) handlePush(rq SendingRequest, chnl *amqp.Channel){
	err := chnl.ExchangeDeclare(rq.id, "fanout", true, false, false, false, nil)
			if err != nil{
				rq.errChnl <- err
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			chnl.PublishWithContext(ctx, rq.id, "", true, true, amqp.Publishing{
				ContentType: "text/plain",
				Body: []byte(rq.msg),
			})
			rq.chnl <- true
}