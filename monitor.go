package bux

import (
	"context"
	"errors"
	"fmt"
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

	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      monitor.GetMaxNumberOfDestinations(),
		OrderByField:  "monitor",
		SortDirection: "desc",
	}

	var destinations []*destinationMonitor
	if err := c.Datastore().GetModels(
		ctx, &[]*Destination{}, conditions, queryParams, &destinations, defaultDatabaseReadTimeout,
	); err != nil && !errors.Is(err, datastore.ErrNoResults) {
		return err
	}

	for _, model := range destinations {
		// todo: skipping the error check?
		_ = monitor.Processor().Add(utils.P2PKHRegexpString, model.LockingScript)
	}

	if c.options.debug && c.Logger() != nil {
		c.Logger().Info(ctx, fmt.Sprintf("[MONITOR] Added %d destinations to monitor", len(destinations)))
	}

	return nil
}
