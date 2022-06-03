package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/logger"
	"github.com/BuxOrg/bux/taskmanager"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithLogger(logger.NewLogger(false, 4)),                                             // Example of using a custom logger
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
