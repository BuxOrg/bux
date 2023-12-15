package taskmanager

import (
	"context"

	taskq "github.com/vmihailenco/taskq/v3"
)

// Tasker is the taskmanager client interface
type Tasker interface {
	RegisterTask(name string, handler interface{}) error
	ResetCron()
	RunTask(ctx context.Context, options *TaskRunOptions) error
	Tasks() map[string]*taskq.Task
	CronJobsInit(cronJobsMap CronJobs) error
	Close(ctx context.Context) error
	Debug(on bool)
	Factory() Factory
	GetTxnCtx(ctx context.Context) context.Context
	IsDebug() bool
	IsNewRelicEnabled() bool
}
