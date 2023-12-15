package taskmanager

import (
	"context"

	taskq "github.com/vmihailenco/taskq/v3"
)

// ClientInterface is the taskmanager client interface
type ClientInterface interface {
	RegisterTask(task *Task) error
	ResetCron()
	RunTask(ctx context.Context, options *TaskOptions) error
	Tasks() map[string]*taskq.Task
	CronJobsInit(cronJobsMap CronJobs) error
	Close(ctx context.Context) error
	Debug(on bool)
	Factory() Factory
	GetTxnCtx(ctx context.Context) context.Context
	IsDebug() bool
	IsNewRelicEnabled() bool
}
