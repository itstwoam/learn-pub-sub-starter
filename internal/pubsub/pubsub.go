package pubsub

import (
	"fmt"
	"encoding/json"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	ctx := context.Background()
	imm := false
	man := false
	msg := amqp.Publishing{ ContentType: "application/json", Body: jsonBytes}
	err = amqp.PublishWithContext(ctx, exchange, key, man, imm, msg)
	if err != nil {
		return err
	}
}
