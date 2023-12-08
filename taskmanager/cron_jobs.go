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
	Name    string
	Period  time.Duration
}

// CronJobsInit registers and runs the cron jobs
func (tm *Client) CronJobsInit(target interface{}, cronJobsList []CronJob) (err error) {
	tm.ResetCron()
	defer func() {
		// stop other, already registered tasks if the func fails
		if err != nil {
			tm.ResetCron()
		}
	}()

	ctx := context.Background()

	for _, taskDef := range cronJobsList {
		if err = tm.RegisterTask(&Task{
			Name:       taskDef.Name,
			RetryLimit: 1,
			Handler: func() error {
				if taskErr := taskDef.Handler(ctx, target); taskErr != nil {
					if tm.options.logger != nil {
						tm.options.logger.Error(ctx, "error running %v task: %v", taskDef.Name, taskErr.Error())
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
			TaskName:       taskDef.Name,
		}); err != nil {
			return
		}
	}
	return
}
