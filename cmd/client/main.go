package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
//	"os/signal"
//	"syscall"
	"github.com/itstwoam/learn-pub-sub-starter/internal/pubsub"
	"github.com/itstwoam/learn-pub-sub-starter/internal/routing"
	"github.com/itstwoam/learn-pub-sub-starter/internal/gamelogic"
)

func main() {
	fmt.Println("Starting Peril client...")
	conString := "amqp://guest:guest@localhost:5672/"
	gameCon, err := amqp.Dial(conString)
	if err != nil {
		fmt.Println("error establishing connection to server, shutting down...")
		os.Exit(1)
	}
	defer gameCon.Close()

	fmt.Println("Connection successful!")

	publishCh, err := gameCon.Channel()
	if err != nil {
		fmt.Println("could not create channel: ", err)
	}

	uName, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Error while getting user name: ", err)
		os.Exit(1)
	}

	myGameState := gamelogic.NewGameState(uName)
	_, _, err = pubsub.DeclareAndBind(gameCon, routing.ExchangePerilDirect, routing.PauseKey+"."+myGameState.GetUsername(), routing.PauseKey, pubsub.Transient)
	//decChan, decQueue, err := pubsub.DeclareAndBind(gameCon, routing.ExhangePerilDirect, routing.PauseKey+"."+uName, routing.PauseKey, pubsub.Transient)
	if err != nil {
		fmt.Println("error binding queue: ", err)
		os.Exit(1)
	}

	pubsub.SubscribeJSON(gameCon, routing.ExchangePerilDirect, routing.PauseKey+"."+myGameState.GetUsername(), routing.PauseKey, pubsub.Transient, handlerPause(myGameState))

	err = pubsub.SubscribeJSON(gameCon, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+myGameState.Player.Username, routing.ArmyMovesPrefix + ".*", pubsub.Transient, handlerMove(myGameState))
	if err != nil {
		fmt.Println("Error when subscribing to army_moves: ", err)
		fmt.Println("Exiting")
		os.Exit(1)
	}

	loop:
	for {
		input := gamelogic.GetInput()
		switch input[0]{
		case "spawn":
			err = myGameState.CommandSpawn(input)
			if err != nil {
				fmt.Errorf("Invalid usage:",err)
			}
		case "move":
			myMove, err := myGameState.CommandMove(input)
			if err != nil {
				fmt.Errorf("Invalid usage:", err)
				continue
			} 
				err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+myGameState.GetUsername(), myMove)
				if err != nil {
					fmt.Println("Error when publishing move: ", err)
					fmt.Println("Failed to publish move.")
				}
				fmt.Println("Published move successfully");
			
		case "status":
			myGameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break loop
		default:
			fmt.Println("I do not understand that command")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState){
	return func(ps routing.PlayingState){
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}
