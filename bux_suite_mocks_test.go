package bux

import (
	"context"

	"github.com/BuxOrg/bux/taskmanager"
	"github.com/vmihailenco/taskq/v3"
)

// taskManagerMock is a base for an empty task manager
type taskManagerMockBase struct{}

func (tm *taskManagerMockBase) Info(context.Context, string, ...interface{}) {}

func (tm *taskManagerMockBase) RegisterTask(*taskmanager.Task) error {
	return nil
}

func (tm *taskManagerMockBase) ResetCron() {}

func (tm *taskManagerMockBase) RunTask(context.Context, *taskmanager.TaskOptions) error {
	return nil
}

func (tm *taskManagerMockBase) Tasks() map[string]*taskq.Task {
	return nil
}

func (tm *taskManagerMockBase) Close(context.Context) error {
	return nil
}

func (tm *taskManagerMockBase) Debug(bool) {}

func (tm *taskManagerMockBase) Engine() taskmanager.Engine {
	return taskmanager.Empty
}

func (tm *taskManagerMockBase) Factory() taskmanager.Factory {
	return taskmanager.FactoryEmpty
}

func (tm *taskManagerMockBase) GetTxnCtx(ctx context.Context) context.Context {
	return ctx
}

func (tm *taskManagerMockBase) IsDebug() bool {
	return false
}

func (tm *taskManagerMockBase) IsNewRelicEnabled() bool {
	return false
}
