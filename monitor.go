package bux

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
)

type destinationMonitor struct {
	LockingScript string `json:"locking_script" toml:"locking_script" yaml:"locking_script" bson:"locking_script"`
}

func (c *Client) loadMonitoredDestinations(ctx context.Context, monitor chainstate.MonitorService) error {
	conditions := map[string]interface{}{
		"monitor": map[string]interface{}{
			"$gt": time.Now().Add(time.Duration(-24*monitor.GetMonitorDays()) * time.Hour),
		},
	}
	var destinations []*destinationMonitor
	err := c.Datastore().GetModels(ctx, &[]*Destination{}, conditions, monitor.GetMaxNumberOfDestinations(), 0, "", "", &destinations, 30*time.Second)
	if err != nil && !errors.Is(err, datastore.ErrNoResults) {
		return err
	}

	for _, model := range destinations {
		fmt.Printf("Monitor: %s", model.LockingScript)
		monitor.Processor().Add(model.LockingScript)
	}

	return nil
}
