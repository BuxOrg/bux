package main

import (
	"context"
	"log"
	"time"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/taskmanager"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithTaskQ(taskmanager.DefaultTaskQConfig("test_queue"), taskmanager.FactoryMemory), // Tasks
		bux.WithCustomCronJobs(func(jobs taskmanager.CronJobs) taskmanager.CronJobs {
			// update the period of the incoming transaction job
			if incommingTransactionJob, ok := jobs[bux.CronJobNameIncomingTransaction]; ok {
				incommingTransactionJob.Period = 10 * time.Second
				jobs[bux.CronJobNameIncomingTransaction] = incommingTransactionJob
			}

			// remove the sync transaction job
			delete(jobs, bux.CronJobNameSyncTransactionSync)

			// add custom job
			jobs["custom_job"] = taskmanager.CronJob{
				Period: 2 * time.Second,
				Handler: bux.BuxClientHandler(func(ctx context.Context, client *bux.Client) error {
					log.Println("custom job!")
					return nil
				}),
			}

			return jobs
		}),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	// wait for the custom cron job to run at least once
	time.Sleep(4 * time.Second)

	log.Println("client loaded!", client.UserAgent())
}
