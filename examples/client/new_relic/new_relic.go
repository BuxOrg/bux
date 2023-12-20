package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/tester"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func main() {
	// EXAMPLE: new relic application
	// replace this with your ALREADY EXISTING new relic application
	app, err := tester.GetNewRelicApp("test-app")
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	var client bux.ClientInterface
	client, err = bux.NewClient(
		newrelic.NewContext(context.Background(), app.StartTransaction("test-txn")), // Set context
		bux.WithNewRelic(app), // New relic application (from your own application or server)
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
