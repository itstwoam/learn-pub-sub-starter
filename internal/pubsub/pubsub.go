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
	gameQueue, err := gameChan.QueueDeclare(queueName, queueType == Durable, queueType == Transient, queueType == Transient, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = gameChan.QueueBind(queueName, key, exchange, false, nil)	
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return gameChan, gameQueue, nil
}
