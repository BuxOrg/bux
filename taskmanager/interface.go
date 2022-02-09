package taskmanager

import (
	"context"

	"github.com/vmihailenco/taskq/v3"
)

// Logger is the logger interface for debug messages
type Logger interface {
	Info(ctx context.Context, message string, params ...interface{})
}

// TaskService is the task related methods
type TaskService interface {
	RegisterTask(task *Task) error
	ResetCron()
	RunTask(ctx context.Context, options *TaskOptions) error
	Tasks() map[string]*taskq.Task
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
