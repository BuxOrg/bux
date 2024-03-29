package main

import (
	"context"
	"log"
	"os"

	"github.com/BuxOrg/bux"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithDebugging(),  // Enable debugging (verbose logs)
		bux.WithEncryption(os.Getenv("BUX_ENCRYPTION_KEY")), // Encryption key for external public keys (paymail)
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent(), "debugging: ", client.IsDebug())
}
