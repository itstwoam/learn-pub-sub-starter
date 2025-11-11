package main

import (
	"fmt"
	"log"
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
		log.Fatalf("error establishing connection to server, shutting down...")
	}
	defer gameCon.Close()

	fmt.Println("Connection successful!")

	publishCh, err := gameCon.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	uName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error while getting user name: %v", err)
	}

	myGameState := gamelogic.NewGameState(uName)
	_, _, err = pubsub.DeclareAndBind(gameCon, routing.ExchangePerilDirect, routing.PauseKey+"."+myGameState.GetUsername(), routing.PauseKey, pubsub.Transient)
	//decChan, decQueue, err := pubsub.DeclareAndBind(gameCon, routing.ExhangePerilDirect, routing.PauseKey+"."+uName, routing.PauseKey, pubsub.Transient)
	if err != nil {
		log.Fatalf("error binding queue: %v", err)
	}

	err = pubsub.SubscribeJSON(gameCon, routing.ExchangePerilDirect, routing.PauseKey+"."+myGameState.GetUsername(), routing.PauseKey, pubsub.Transient, handlerPause(myGameState))
	

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(gameCon, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+myGameState.Player.Username, routing.ArmyMovesPrefix + ".*", pubsub.Transient, handlerMove(myGameState, publishCh))
	if err != nil {
		fmt.Println("Error when subscribing to army_moves: ", err)
		fmt.Println("Exiting")
		os.Exit(1)
	}

	err = pubsub.SubscribeJSON(
		gameCon, routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(myGameState),
		)

	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

	loop:
	for {
		fmt.Println("Awaiting input in main loop")
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
			return
		default:
			fmt.Println("I do not understand that command")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType{
	return func(ps routing.PlayingState) pubsub.AckType{
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	 	return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType{
		defer fmt.Println("> ")
		mo := gs.HandleMove(move)
		switch mo {
		case gamelogic.MoveOutcomeSamePlayer:
			fallthrough
		case gamelogic.MoveOutcomeSafe:
			fmt.Println("--done--")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
				)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			fmt.Println("--done--")
			return pubsub.Ack
		}
		fmt.Println("error: unkown move outcome")
			fmt.Println("--done--")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Println("> ")
		warOutcome, _, _ := gs.HandleWar(dw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			fmt.Println("--done--")
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			fmt.Println("--done--")
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			fmt.Println("--done--")
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			fmt.Println("--done--")
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			fmt.Println("--done--")
			return pubsub.Ack
		}
		
		fmt.Println("error: unkown war outcome")
			fmt.Println("--done--")
		return pubsub.NackDiscard
	}
}
