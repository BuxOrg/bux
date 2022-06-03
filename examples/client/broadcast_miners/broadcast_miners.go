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

	// Create a custom miner (using your api key for custom rates)
	miners, _ := minercraft.DefaultMiners()
	minerTaal := minercraft.MinerByName(miners, minercraft.MinerTaal)
	minerTaal.Token = os.Getenv("BUX_TAAL_API_KEY")

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

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
