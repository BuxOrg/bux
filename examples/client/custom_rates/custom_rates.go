package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/BuxOrg/bux"
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
					URL:   "https://tapi.taal.com/arc",
					Type:  minercraft.Arc,
				},
			},
		},
	}

	// Create the client
	client, err := bux.NewClient(
		context.Background(),                   // Set context
		bux.WithAutoMigrate(bux.BaseModels...), // All models
		bux.WithMinercraftAPIs(minerCraftApis),
		bux.WithArc(),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

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
