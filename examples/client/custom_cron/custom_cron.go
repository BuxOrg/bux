package main

import (
	"context"
	"log"
	"time"

	"github.com/BuxOrg/bux"
)

func main() {
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithCronCustmPeriod(bux.CronJobNameDraftTransactionCleanUp, 2*time.Second),
		bux.WithCronCustmPeriod(bux.CronJobNameSyncTransactionSync, 4*time.Second),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	// wait for the customized cron jobs to run at least once
	time.Sleep(8 * time.Second)

	log.Println("client loaded!", client.UserAgent())
}
