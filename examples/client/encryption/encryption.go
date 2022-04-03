package main

import (
	"context"
	"log"
	"os"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/taskmanager"
)

func main() {
	client, err := bux.NewClient(
		context.Background(),                                                                   // Set context
		bux.WithFreeCache(),                                                                    // Cache
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithDebugging(),                                                                    // Enable debugging (verbose logs)
		bux.WithEncryption(os.Getenv("BUX_ENCRYPTION_KEY")),                                    // Encryption key for external public keys (paymail)
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	log.Println("client loaded!", client.UserAgent(), "debugging: ", client.IsDebug())
}
