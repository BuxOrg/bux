package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/taskmanager"
)

type cronFunc func(ctx context.Context, client *Client) error

type cronDefinition struct {
	Handler cronFunc
	Name    string
	Period  time.Duration
}

func registerCronTasks(client *Client, definitions []cronDefinition) (err error) {
	tm := client.Taskmanager()
	tm.ResetCron()
	defer func() {
		// stop other, already registered tasks if the func fails
		if err != nil {
			tm.ResetCron()
		}
	}()

	ctx := context.Background()

	for _, taskDef := range definitions {
		if err = tm.RegisterTask(&taskmanager.Task{
			Name:       taskDef.Name,
			RetryLimit: 1,
			Handler: func() error {
				if taskErr := taskDef.Handler(ctx, client); taskErr != nil {
					client.Logger().Error(ctx, "error running %v task: %v", taskDef.Name, taskErr.Error())
				}
				return nil
			},
		}); err != nil {
			return
		}

		// Run the task periodically
		if err = tm.RunTask(ctx, &taskmanager.TaskOptions{
			RunEveryPeriod: taskDef.Period,
			TaskName:       taskDef.Name,
		}); err != nil {
			return
		}
	}
	return
}
