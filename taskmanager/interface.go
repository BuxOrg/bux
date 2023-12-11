package taskmanager

import (
	"context"

	taskq "github.com/vmihailenco/taskq/v3"
)

// TaskService is the task related methods
type TaskService interface {
	RegisterTask(task *Task) error
	ResetCron()
	RunTask(ctx context.Context, options *TaskOptions) error
	Tasks() map[string]*taskq.Task
	CronJobsInit(target interface{}, cronJobsList map[string]CronJob) error
}

// CronService is the cron service provider
type CronService interface {
	AddFunc(spec string, cmd func()) (int, error)
	New()
	Start()
	Stop()
}

// ClientInterface is the taskmanager client interface
type ClientInterface interface {
	TaskService
	Close(ctx context.Context) error
	Debug(on bool)
	Engine() Engine
	Factory() Factory
	GetTxnCtx(ctx context.Context) context.Context
	IsDebug() bool
	IsNewRelicEnabled() bool
}
