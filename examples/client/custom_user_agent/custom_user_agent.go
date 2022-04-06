package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/taskmanager"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithUserAgent("my-custom-user-agent"),                                              // Custom user agent
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	log.Println("client loaded!", client.UserAgent())
}
