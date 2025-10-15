package pubsub

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"context"
	"fmt"
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
