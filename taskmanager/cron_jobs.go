package taskmanager

import (
	"context"
	"time"
)

// CronJobHandler is the handler for a cron job
type CronJobHandler func(ctx context.Context, target interface{}) error

// CronJob definition, params reduced to the minimum, all required
type CronJob struct {
	Handler CronJobHandler
	Period  time.Duration
}

// CronJobsInit registers and runs the cron jobs
func (tm *Client) CronJobsInit(target interface{}, cronJobsList map[string]CronJob) (err error) {
	tm.ResetCron()
	defer func() {
		// stop other, already registered tasks if the func fails
		if err != nil {
			tm.ResetCron()
		}
	}()

	ctx := context.Background()

	for name, taskDef := range cronJobsList {
		handler := taskDef.Handler
		if err = tm.RegisterTask(&Task{
			Name:       name,
			RetryLimit: 1,
			Handler: func() error {
				if taskErr := handler(ctx, target); taskErr != nil {
					if tm.options.logger != nil {
						tm.options.logger.Error(ctx, "error running %v task: %v", name, taskErr.Error())
					}
				}
				return nil
			},
		}); err != nil {
			return
		}

		// Run the task periodically
		if err = tm.RunTask(ctx, &TaskOptions{
			RunEveryPeriod: taskDef.Period,
			TaskName:       name,
		}); err != nil {
			return
		}
	}
	return
}
