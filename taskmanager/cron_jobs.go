package taskmanager

import (
	"context"
	"time"
)

// CronJobHandler is the handler for a cron job
type CronJobHandler func(ctx context.Context) error

// CronJob definition, params reduced to the minimum, all required
type CronJob struct {
	Handler CronJobHandler
	Period  time.Duration
}

// CronJobs as a map prevents duplicate jobs with the same name
type CronJobs map[string]CronJob

// CronJobsInit registers and runs the cron jobs
func (tm *Client) CronJobsInit(cronJobsMap CronJobs) (err error) {
	tm.ResetCron()
	defer func() {
		// stop other, already registered tasks if the func fails
		if err != nil {
			tm.ResetCron()
		}
	}()

	ctx := context.Background()

	for name, taskDef := range cronJobsMap {
		handler := taskDef.Handler
		if err = tm.RegisterTask(&Task{
			Name:       name,
			RetryLimit: 1,
			Handler: func() error {
				if taskErr := handler(ctx); taskErr != nil {
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
