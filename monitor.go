package bux

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BuxOrg/bux/chainstate"
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
		client.Logger().Info(ctx, fmt.Sprintf("[MONITOR] Added %d destinations to monitor", len(destinations)))
	}

	return nil
}

// checkMonitorHeartbeat will check for a Monitor heartbeat key and detect if it's locked or not
func checkMonitorHeartbeat(ctx context.Context, client ClientInterface, lockID string) (bool, error) {

	// Make sure cachestore is loaded (safety check)
	cs := client.Cachestore()
	if cs == nil {
		return false, nil
	}

	// Check if there is already a monitor loaded using this unique lock id & detect last heartbeat
	lastHeartBeatString, _ := cs.Get(ctx, fmt.Sprintf(lockKeyMonitorLockID, lockID))
	if len(lastHeartBeatString) > 0 { // Monitor has already loaded with this LockID

		currentUnixTime := time.Now().UTC().Unix()

		// Convert the heartbeat and then compare
		unixTime, err := strconv.Atoi(lastHeartBeatString)
		if err != nil {
			return false, err
		}

		// Heartbeat is good - skip loading the monitor
		if int64(unixTime+defaultMonitorHeartbeatMax) > currentUnixTime {
			client.Logger().Info(ctx, fmt.Sprintf("monitor has already been loaded using this lockID: %s last heartbeat: %d", lockID, unixTime))
			return true, nil
		}
		client.Logger().Info(ctx, fmt.Sprintf("found monitor lockID: %s but heartbeat is out of range: %d vs %d", lockID, unixTime, currentUnixTime))
	}
	return false, nil
}

// startDefaultMonitor will create a handler, start monitor, and store the first heartbeat
func startDefaultMonitor(ctx context.Context, client ClientInterface, monitor chainstate.MonitorService) error {

	// Create a handler and load destinations if option has been set
	handler := NewMonitorHandler(ctx, client, monitor)
	if client.Chainstate().Monitor().LoadMonitoredDestinations() {
		if err := loadMonitoredDestinations(ctx, client, monitor); err != nil {
			return err
		}
	}

	// Start the monitor
	if err := monitor.Start(ctx, &handler); err != nil {
		return err
	}

	// Set the cache-key lock for this monitor (with a heartbeat time of now)
	if len(monitor.GetLockID()) > 0 {
		return client.Cachestore().Set(
			ctx,
			fmt.Sprintf(lockKeyMonitorLockID, monitor.GetLockID()),
			fmt.Sprintf("%d", time.Now().UTC().Unix()),
		)
	}
	return nil
}
