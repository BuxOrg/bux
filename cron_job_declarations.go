package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/taskmanager"
)

// Cron job names to be used in WithCronCustomPeriod
const (
	CronJobNameDraftTransactionCleanUp  = "draft_transaction_clean_up"
	CronJobNameIncomingTransaction      = "incoming_transaction_process"
	CronJobNameSyncTransactionBroadcast = "sync_transaction_broadcast"
	CronJobNameSyncTransactionSync      = "sync_transaction_sync"
)

type cronJobHandler func(ctx context.Context, client *Client) error

// here is where we define all the cron jobs for the client
func (c *Client) cronJobs() taskmanager.CronJobs {
	// handler adds the client pointer to the cronJobTask by using a closure
	handler := func(cronJobTask cronJobHandler) taskmanager.CronJobHandler {
		return func(ctx context.Context) error {
			return cronJobTask(ctx, c)
		}
	}

	return taskmanager.CronJobs{
		CronJobNameDraftTransactionCleanUp: {
			Period:  60 * time.Second,
			Handler: handler(taskCleanupDraftTransactions),
		},
		CronJobNameIncomingTransaction: {
			Period:  30 * time.Second,
			Handler: handler(taskProcessIncomingTransactions),
		},
		CronJobNameSyncTransactionBroadcast: {
			Period:  30 * time.Second,
			Handler: handler(taskBroadcastTransactions),
		},
		CronJobNameSyncTransactionSync: {
			Period:  120 * time.Second,
			Handler: handler(taskSyncTransactions),
		},
	}
}
