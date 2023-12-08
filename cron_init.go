package bux

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/taskmanager"
)

type cronHandler func(ctx context.Context, client *Client) error

type cronJob struct {
	Handler cronHandler
	Name    string
	Period  time.Duration
}

func (c *Client) cronInit(cronJobsList []cronJob) (err error) {
	tm := c.Taskmanager()
	tm.ResetCron()
	defer func() {
		// stop other, already registered tasks if the func fails
		if err != nil {
			tm.ResetCron()
		}
	}()

	ctx := context.Background()

	for _, taskDef := range cronJobsList {
		if err = tm.RegisterTask(&taskmanager.Task{
			Name:       taskDef.Name,
			RetryLimit: 1,
			Handler: func() error {
				if taskErr := taskDef.Handler(ctx, c); taskErr != nil {
					c.Logger().Error(ctx, "error running %v task: %v", taskDef.Name, taskErr.Error())
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
