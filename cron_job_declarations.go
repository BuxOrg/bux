package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/taskmanager"
)

// Cron job names to be used in WithCronCustomPeriod
const (
	CronJobNameDraftTransactionCleanUp  = "draft_transaction_clean_up"
	CronJobNameSyncTransactionBroadcast = "sync_transaction_broadcast"
	CronJobNameSyncTransactionSync      = "sync_transaction_sync"
	CronJobNameCalculateMetrics         = "calculate_metrics"
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

	jobs := taskmanager.CronJobs{
		CronJobNameDraftTransactionCleanUp: {
			Period:  60 * time.Second,
			Handler: handler(taskCleanupDraftTransactions),
		},
		CronJobNameSyncTransactionBroadcast: {
			Period:  2 * time.Minute,
			Handler: handler(taskBroadcastTransactions),
		},
		CronJobNameSyncTransactionSync: {
			Period:  5 * time.Minute,
			Handler: handler(taskSyncTransactions),
		},
	}

	if _, enabled := c.Metrics(); enabled {
		jobs[CronJobNameCalculateMetrics] = taskmanager.CronJob{
			Period:  15 * time.Second,
			Handler: handler(taskCalculateMetrics),
		}
	}

	return jobs
}
