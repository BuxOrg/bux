package bux

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
)

func (c *Client) loadMonitoredDestinations(ctx context.Context, monitor chainstate.MonitorService) error {
	var models []*Destination
	conditions := map[string]interface{}{
		"monitorConfig": map[string]interface{}{
			"$gt": time.Now().Add(time.Duration(-24*monitor.GetMonitorDays()) * time.Hour),
		},
	}
	projection := []string{"locking_script"}
	err := c.Datastore().GetModels(ctx, &models, conditions, monitor.GetMaxNumberOfDestinations(), 0, "", "", &projection, 30*time.Second)
	if err != nil && !errors.Is(err, datastore.ErrNoResults) {
		return err
	}

	for _, model := range models {
		fmt.Printf("Monitor: %s", model.LockingScript)
		monitor.Add(model.LockingScript)
	}

	return nil
}
