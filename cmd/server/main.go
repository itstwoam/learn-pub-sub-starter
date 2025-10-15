package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
	"github.com/itstwoam/learn-pubsub-starter/internal/pubsub"
)

func main() {
	fmt.Println("Starting Peril server...")
	conString := "amqp://guest:guest@localhost:5672/"
	gameCon, err := amqp.Dial(conString)
	if err != nil {
		fmt.Println("error establishing connection to server, shutting down...")
		os.Exit(1)
	}
	defer gameCon.Close()

	fmt.Println("Connection successful!")

	gameChan, err := gameCon.Channel()
	if err != nil {
		fmt.Prinln("error creating amqp channel")
		os.Exit(1)
	}
	
	err = pubsub.PublishJSON(gameChan, routing.ExchangerPerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})	
	if err != nil {
		fmt.Println("error in pubsub.PublishJSON()")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		_ = <-sigs
		fmt.Println("\nGracefully shutting down...")
		gameCon.Close()
		os.Exit(0)
	
}
