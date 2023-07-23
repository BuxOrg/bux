package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/tonicpow/go-minercraft/v2"
)

func main() {

	const testXPub = "xpub661MyMwAqRbcFrBJbKwBGCB7d3fr2SaAuXGM95BA62X41m6eW2ehRQGW4xLi9wkEXUGnQZYxVVj4PxXnyrLk7jdqvBAs1Qq9gf6ykMvjR7J"

	// Create a custom miner (using your api key for custom rates)
	miners, _ := minercraft.DefaultMiners()
	minerTaal := minercraft.MinerByName(miners, minercraft.MinerTaal)
	minerCraftApis := []*minercraft.MinerAPIs{
		{
			MinerID: minerTaal.MinerID,
			APIs: []minercraft.API{
				{
					Token: os.Getenv("BUX_TAAL_API_KEY"),
					URL: "https://tapi.taal.com/arc",
					Type: minercraft.Arc,
				},
			},
		},
	}

	// Create the client
	client, err := bux.NewClient(
		context.Background(),                   // Set context
		bux.WithAutoMigrate(bux.BaseModels...), // All models
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithBroadcastMiners([]*chainstate.Miner{{Miner: minerTaal}}),                       // This will auto-fetch a policy using the token (api key)
		bux.WithMinercraftAPIs(minerCraftApis),
		bux.WithArc(),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	// Get the miners
	broadcastMiners := client.Chainstate().BroadcastMiners()
	for _, miner := range broadcastMiners {
		log.Println("miner", miner.Miner)
		log.Println("fee", miner.FeeUnit)
		log.Println("last_checked", miner.FeeLastChecked.String())
	}

	// Create an xPub
	var xpub *bux.Xpub
	if xpub, err = client.NewXpub(
		context.Background(),
		testXPub,
	); err != nil {
		log.Fatalln("error: " + err.Error())
	}

	// Create a draft transaction
	var draft *bux.DraftTransaction
	draft, err = client.NewTransaction(context.Background(), xpub.RawXpub(), &bux.TransactionConfig{
		ExpiresIn: 10 * time.Second,
		SendAllTo: &bux.TransactionOutput{To: "mrz@moneybutton.com"},
	})
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	// Custom fee
	log.Println("fee unit", draft.Configuration.FeeUnit)
}
