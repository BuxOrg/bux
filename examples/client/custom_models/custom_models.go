package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithDebugging(),
		bux.WithAutoMigrate(bux.BaseModels...),
		bux.WithModels(NewExample("example-field")), // Add additional custom models to Bux
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
