package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	//"os/signal"
	//"syscall"
	"github.com/itstwoam/learn-pub-sub-starter/internal/pubsub"
	"github.com/itstwoam/learn-pub-sub-starter/internal/routing"
	"github.com/itstwoam/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	fmt.Println("Starting Peril server...")
	playState := routing.PlayingState{ IsPaused: false}
	conString := "amqp://guest:guest@localhost:5672/"
	gameCon, err := amqp.Dial(conString)
	if err != nil {
		fmt.Println("error establishing connection to server, shutting down...")
		os.Exit(1)
	}
	defer gameCon.Close()

	fmt.Println("Connection successful!")

	gamelogic.PrintServerHelp()
	gameChan, err := gameCon.Channel()
	if err != nil {
		fmt.Println("error creating amqp channel")
		os.Exit(1)
	}

	/**amqp.Channel, amqp.Queue, error*/_, _, err = pubsub.DeclareAndBind(gameCon, routing.ExchangePerilTopic, "game_logs", "game_logs.*", pubsub.Durable) 
	loop:
	for {
		input := gamelogic.GetInput()
		switch input[0]{
		case "pause":
			//fmt.Println("exchange:", routing.ExchangePerilDirect, "key:", routing.PauseKey)	
			if err = pubsub.PublishJSON(gameChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
				fmt.Println("pause publish error:", err)
			}else {
				fmt.Println("published pause message")
				playState.IsPaused = true
			}
		case "resume":
			if playState.IsPaused == true {
				if err = pubsub.PublishJSON(gameChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
					fmt.Println("unpause publish error:", err)
				}else {
					fmt.Println("published unpause message")
					playState.IsPaused = false
				}
			}
		case "quit":
			fmt.Println("Quit command recieved, shutting down gracefully")
			break loop
		case "help":
			gamelogic.PrintServerHelp()
		default:
			fmt.Println("I do not understand that command")
		}
	}
}
