package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithDebugging(),  // Enable debugging (verbose logs)
		bux.WithChainstateOptions(true, true, true, true), // Broadcasting enabled by default
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent(), "debugging: ", client.IsDebug())
}
