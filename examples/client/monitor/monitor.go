package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mrz1836/go-datastore"

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
		bux.WithMonitoring(context.Background(), &chainstate.MonitorOptions{
			AuthToken:                   os.Getenv("BUX_MONITOR_AUTH_TOKEN"),
			BuxAgentURL:                 "wss://" + os.Getenv("BUX_MONITOR_URL"),
			Debug:                       true,
			FalsePositiveRate:           0,
			LoadMonitoredDestinations:   true,
			LockID:                      "unique-lock-id-for-multiple-servers",
			MaxNumberOfDestinations:     25000,
			MonitorDays:                 5,
			ProcessMempoolOnConnect:     false,
			ProcessorType:               chainstate.FilterRegex,
			SaveTransactionDestinations: false,
		}),
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithDebugging(), // Enable debugging (verbose logs)
		bux.WithChainstateOptions(true, true, true, true), // Broadcasting enabled by default
		bux.WithAutoMigrate(bux.BaseModels...),
	)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer func() {
		_ = client.Close(context.Background())
	}()

	m := client.Chainstate().Monitor()

	// Create a new handler
	handler := bux.NewMonitorHandler(context.Background(), client, m)

	// Start
	if err = m.Start(context.Background(), &handler, func() {
		// callback when the monitor stops
	}); err != nil {
		log.Fatalf(err.Error())
	}

	// Add a regex filter
	if err = m.Add("006a", ""); err != nil {
		log.Fatal(err.Error())
	}

	// Pause
	time.Sleep(time.Second * 10)

	// Stop the monitor
	if err = m.Stop(context.Background()); err != nil {
		log.Fatalf(err.Error())
	}
	time.Sleep(time.Second * 5)
	log.Println("Complete!")
}
