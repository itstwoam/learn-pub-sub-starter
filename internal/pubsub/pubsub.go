package pubsub

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"context"
	"fmt"
)

type SimpleQueueType string

const (Transient SimpleQueueType = "transient"
	Durable SimpleQueueType = "durable"
)

type AckType string

const (Ack AckType = "ack"
	NackRequeue AckType = "nackR"
	NackDiscard AckType = "nackD"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		fmt.Println("error in PublishJSON() while marshalling JSON")
		return err
	}

	ctx := context.Background()
	imm := false
	man := false
	msg := amqp.Publishing{ ContentType: "application/json", Body: jsonBytes}
	err = ch.PublishWithContext(ctx, exchange, key, man, imm, msg)
	if err != nil {
		fmt.Println("error in PublishJSON() when calling PublishWithContext")
		return err
	}
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, 
	queueName,
	key string,
	queueType SimpleQueueType, 
) (*amqp.Channel, amqp.Queue, error){
	gameChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	gameQueue, err := gameChan.QueueDeclare(queueName, queueType == Durable, queueType == Transient, queueType == Transient, false, args)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = gameChan.QueueBind(queueName, key, exchange, false, nil)	
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return gameChan, gameQueue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	) error {
	//decChannel, decQueue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	decChannel, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	decChan, err := decChannel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go startDeliveryWorker(decChan, handler)
	return nil
}

func startDeliveryWorker[T any](deliveries <-chan amqp.Delivery, handler func(T) AckType){
	go func() {
		for msg := range deliveries {
			var t T
			if err := json.Unmarshal(msg.Body, &t); err != nil {
				fmt.Printf("Failed to unmarshal: %v", err)
				continue
			}
			ack := handler(t)

			switch ack {
			case Ack:
				fmt.Println("Acknowledged")
				msg.Ack(false)
				//_ = msg.Ack(false)
			case NackRequeue:
				fmt.Println("Nacked and requeued")
				//_ = msg.Nack(false, true)
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("Nacked and discarded")
				//_ = msg.Nack(false, false)
				msg.Nack(false, false)
			default:
				fmt.Println("unkown AckType, message not acknowledged")
			}
		}
	}()
}
