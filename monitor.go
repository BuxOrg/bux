package bux

import (
	"context"
	"errors"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
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
		_ = monitor.Processor().Add(utils.P2PKHRegexpString, model.LockingScript)
	}

	return nil
}
