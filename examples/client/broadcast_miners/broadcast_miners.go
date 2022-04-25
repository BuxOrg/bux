package main

import (
	"context"
	"log"
	"os"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/tonicpow/go-minercraft"
)

func main() {

	// Use a custom miner
	minerTaal := &minercraft.Miner{
		MinerID: "030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e",
		Name:    "Taal",
		URL:     "https://merchantapi.taal.com",
		Token:   os.Getenv("BUX_TAAL_API_KEY"), // This is optional - for custom rates
	}

	// Create the client
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithBroadcastMiners([]*chainstate.Miner{{Miner: minerTaal}}),                       // This will auto-fetch a policy using the token (api key)
		bux.WithQueryMiners([]*chainstate.Miner{{Miner: minerTaal}}),                           // This will only use this as a query provider
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	log.Println("client loaded!", client.UserAgent())
}
