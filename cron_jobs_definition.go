package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/taskmanager"
)

var cronJobs = []taskmanager.CronJob{
	{
		Name:    "draft_transaction_clean_up",
		Period:  defaultMonitorHeartbeat * time.Second,
		Handler: buxClientHandler(taskCleanupDraftTransactions),
	},
	{
		Name:    "incoming_transaction_process",
		Period:  30 * time.Second,
		Handler: buxClientHandler(taskProcessIncomingTransactions),
	},
	{
		Name:    "sync_transaction_broadcast",
		Period:  30 * time.Second,
		Handler: buxClientHandler(taskBroadcastTransactions),
	},
	{
		Name:    "sync_transaction_sync",
		Period:  120 * time.Second,
		Handler: buxClientHandler(taskSyncTransactions),
	},
}

// buxClientHandler converts a handler with a *Client target to a generic taskmanager.CronJobHandler
func buxClientHandler(handler func(ctx context.Context, client *Client) error) taskmanager.CronJobHandler {
	return func(ctx context.Context, target interface{}) error {
		return handler(ctx, target.(*Client))
	}
}
