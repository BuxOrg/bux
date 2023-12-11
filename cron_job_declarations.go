package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/taskmanager"
)

const (
	CronJobNameDraftTransactionCleanUp  = "draft_transaction_clean_up"
	CronJobNameIncomingTransaction      = "incoming_transaction_process"
	CronJobNameSyncTransactionBroadcast = "sync_transaction_broadcast"
	CronJobNameSyncTransactionSync      = "sync_transaction_sync"
)

// here is where we define all the cron jobs for the client
var defaultCronJobs = taskmanager.CronJobs{
	CronJobNameDraftTransactionCleanUp: {
		Period:  defaultMonitorHeartbeat * time.Second,
		Handler: BuxClientHandler(taskCleanupDraftTransactions),
	},
	CronJobNameIncomingTransaction: {
		Period:  30 * time.Second,
		Handler: BuxClientHandler(taskProcessIncomingTransactions),
	},
	CronJobNameSyncTransactionBroadcast: {
		Period:  30 * time.Second,
		Handler: BuxClientHandler(taskBroadcastTransactions),
	},
	CronJobNameSyncTransactionSync: {
		Period:  120 * time.Second,
		Handler: BuxClientHandler(taskSyncTransactions),
	},
}

// utility function - converts a handler with the *Client target to a generic taskmanager.CronJobHandler
func BuxClientHandler(handler func(ctx context.Context, client *Client) error) taskmanager.CronJobHandler {
	return func(ctx context.Context, target interface{}) error {
		return handler(ctx, target.(*Client))
	}
}
