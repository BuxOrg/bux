package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/logging"
	"github.com/bitcoin-sv/go-broadcast-client/broadcast"
	broadcastclient "github.com/bitcoin-sv/go-broadcast-client/broadcast/broadcast-client"
)

func buildBroadcastClient() broadcast.Client {
	logger := logging.GetDefaultLogger()
	builder := broadcastclient.Builder().WithArc(
		broadcastclient.ArcClientConfig{
			APIUrl: "https://tapi.taal.com/arc",
			Token:  os.Getenv("BUX_TAAL_API_KEY"),
		},
		logger,
	)

	return builder.Build()
}

func main() {
	ctx := context.Background()
	const testXPub = "xpub661MyMwAqRbcFrBJbKwBGCB7d3fr2SaAuXGM95BA62X41m6eW2ehRQGW4xLi9wkEXUGnQZYxVVj4PxXnyrLk7jdqvBAs1Qq9gf6ykMvjR7J"

	client, err := bux.NewClient(
		ctx,
		bux.WithAutoMigrate(bux.BaseModels...),
		bux.WithBroadcastClient(buildBroadcastClient()),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer client.Close(ctx)

	xpub, err := client.NewXpub(ctx, testXPub)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	draft, err := client.NewTransaction(ctx, xpub.RawXpub(), &bux.TransactionConfig{
		ExpiresIn: 10 * time.Second,
		SendAllTo: &bux.TransactionOutput{To: os.Getenv("BUX_MY_PAYMAIL")},
	})
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	// Custom fee
	log.Println("fee unit", draft.Configuration.FeeUnit)
}
