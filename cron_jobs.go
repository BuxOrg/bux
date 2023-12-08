package bux

import (
	"context"
	"time"
)

var cronJobs = []cronJob{
	{
		Name:   "draft_transaction_clean_up",
		Period: defaultMonitorHeartbeat * time.Second,
		Handler: func(ctx context.Context, client *Client) error {
			return taskCleanupDraftTransactions(ctx, client.Logger(), WithClient(client))
		},
	},
	{
		Name:   "incoming_transaction_process",
		Period: 30 * time.Second,
		Handler: func(ctx context.Context, client *Client) error {
			return taskProcessIncomingTransactions(ctx, client.Logger(), WithClient(client))
		},
	},
	{
		Name:   "sync_transaction_broadcast",
		Period: 30 * time.Second,
		Handler: func(ctx context.Context, client *Client) error {
			return taskBroadcastTransactions(ctx, client.Logger(), WithClient(client))
		},
	},
	{
		Name:   "sync_transaction_sync",
		Period: 120 * time.Second,
		Handler: func(ctx context.Context, client *Client) error {
			return taskSyncTransactions(ctx, client, WithClient(client))
		},
	},
}
