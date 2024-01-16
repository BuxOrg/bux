package main

import (
	"context"
	"log"
	"os"

	"github.com/BuxOrg/bux"
	"github.com/tonicpow/go-minercraft/v2"
)

func main() {
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
		context.Background(), // Set context
		bux.WithMinercraftAPIs(minerCraftApis),
		bux.WithArc(),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
