package main

import (
	"context"
	"log"
	"os"

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

	client, err := bux.NewClient(
		ctx,
		bux.WithBroadcastClient(buildBroadcastClient()),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer client.Close(ctx)

	log.Println("client loaded!", client.UserAgent())
}
