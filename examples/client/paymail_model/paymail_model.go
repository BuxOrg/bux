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
		bux.WithFreeCache(),  // Cache
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithAutoMigrate(&bux.PaymailAddress{
			Model: *bux.NewBaseModel(bux.ModelPaymail), // Add the paymail model (for use in BUX server & clients)
		}),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	log.Println("client loaded!", client.UserAgent())
}
