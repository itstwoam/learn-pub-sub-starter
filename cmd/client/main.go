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

	uName, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Error while getting user name: ", err)
		os.Exit(1)
	}

	_, _, err = pubsub.DeclareAndBind(gameCon, routing.ExchangePerilDirect, routing.PauseKey+"."+uName, routing.PauseKey, pubsub.Transient)
	//decChan, decQueue, err := pubsub.DeclareAndBind(gameCon, routing.ExhangePerilDirect, routing.PauseKey+"."+uName, routing.PauseKey, SimpleQueueType.Transient)
	if err != nil {
		fmt.Println("error binding queue: ", err)
		os.Exit(1)
	}
	_ = gamelogic.GetInput()
}
