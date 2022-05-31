package main

import (
	"context"
	"log"
	"time"

	"github.com/BuxOrg/bux/datastore"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/taskmanager"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithSQLite(&datastore.SQLiteConfig{ // Load using a sqlite configuration
			CommonConfig: datastore.CommonConfig{
				Debug:                 false,
				MaxConnectionIdleTime: 10 * time.Second,
				MaxConnectionTime:     10 * time.Second,
				MaxIdleConnections:    10,
				MaxOpenConnections:    10,
				TablePrefix:           "bux",
			},
		}),

		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithDebugging(), // Enable debugging (verbose logs)
		bux.WithChainstateOptions(true, true, true, true), // Broadcasting enabled by default
		bux.WithAutoMigrate(bux.BaseModels...),
	)
	if err != nil {
		log.Fatalf(err.Error())
	}

	m := chainstate.NewMonitor(context.Background(), &chainstate.MonitorOptions{
		BuxAgentURL:             "wss://bux-agent.siftbitcoin.com/websocket",
		AuthToken:               "TOKEN",
		ProcessorType:           "regex",
		ProcessMempoolOnConnect: false,
	})

	handler := bux.NewMonitorHandler(context.Background(), client, m)
	err = m.Start(&handler)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = m.Add("006a", "")
	if err != nil {
		log.Fatal(err.Error())
	}
	time.Sleep(time.Second * 20)
	err = m.Stop()
	if err != nil {
		log.Fatalf(err.Error())
	}
	time.Sleep(time.Minute * 20)
}
