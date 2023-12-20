package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux/logging"

	"github.com/BuxOrg/bux"
)

func main() {
	client, err := bux.NewClient(
		context.Background(),                       // Set context
		bux.WithLogger(logging.GetDefaultLogger()), // Example of using a custom logger
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
