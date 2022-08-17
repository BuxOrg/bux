package bux

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/cluster"
	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-datastore"
)

// destinationMonitor is the struct of responses for Monitoring
type destinationMonitor struct {
	LockingScript string `json:"locking_script" toml:"locking_script" yaml:"locking_script" bson:"locking_script"`
}

// loadMonitoredDestinations will load destinations that should be monitored
func loadMonitoredDestinations(ctx context.Context, client ClientInterface, monitor chainstate.MonitorService) error {

	// Create conditions using the max monitor days
	conditions := map[string]interface{}{
		"monitor": map[string]interface{}{
			"$gt": time.Now().Add(time.Duration(-24*monitor.GetMonitorDays()) * time.Hour),
		},
	}

	// Create monitor query with max destinations
	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      monitor.GetMaxNumberOfDestinations(),
		OrderByField:  "monitor",
		SortDirection: "desc",
	}

	// Get all destinations that match the query
	var destinations []*destinationMonitor
	if err := client.Datastore().GetModels(
		ctx, &[]*Destination{}, conditions, queryParams, &destinations, defaultDatabaseReadTimeout,
	); err != nil && !errors.Is(err, datastore.ErrNoResults) {
		return err
	}

	// Loop all destinations and add to Monitor
	for _, model := range destinations {
		if err := monitor.Processor().Add(utils.P2PKHRegexpString, model.LockingScript); err != nil {
			return err
		}
	}

	// Debug line
	if client.IsDebug() && client.Logger() != nil {
		client.Logger().Info(ctx, fmt.Sprintf(
			"[MONITOR] Added %d destinations to monitor with hash %s",
			len(destinations), monitor.Processor().GetHash(),
		))
	}

	return nil
}

// startDefaultMonitor will create a handler, start monitor, and store the first heartbeat
func startDefaultMonitor(ctx context.Context, client ClientInterface, monitor chainstate.MonitorService) error {

	if client.Chainstate().Monitor().LoadMonitoredDestinations() {
		if err := loadMonitoredDestinations(ctx, client, monitor); err != nil {
			return err
		}
	}

	_, err := client.Cluster().Subscribe(cluster.DestinationNew, func(data string) {
		if monitor.IsDebug() {
			monitor.Logger().Info(ctx, fmt.Sprintf("[MONITOR] added %s destination to monitor: %s", utils.P2PKHRegexpString, data))
		}
		if err := monitor.Processor().Add(utils.P2PKHRegexpString, data); err != nil {
			client.Logger().Error(ctx, "could not add destination to monitor")
		}
	})
	if err != nil {
		return err
	}

	if monitor.IsDebug() {
		// capture keyboard input and allow start and stop of the monitor
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				text := scanner.Text()
				fmt.Printf("KEYBOARD input: %s\n", text)
				if text == "e" {
					if err = monitor.Stop(ctx); err != nil {
						fmt.Printf("ERROR: %s\n", err.Error())
					}
				}
			}
		}()
	}

	return nil
}
