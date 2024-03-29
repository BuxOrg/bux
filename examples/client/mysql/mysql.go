package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/BuxOrg/bux"
	"github.com/mrz1836/go-datastore"
)

func main() {
	defaultTimeouts := 10 * time.Second

	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithSQL(datastore.MySQL, &datastore.SQLConfig{ // Load using a MySQL configuration
			CommonConfig: datastore.CommonConfig{
				Debug:                 true,
				MaxConnectionIdleTime: defaultTimeouts,
				MaxConnectionTime:     defaultTimeouts,
				MaxIdleConnections:    10,
				MaxOpenConnections:    10,
				TablePrefix:           "bux",
			},
			Driver:    datastore.MySQL.String(),
			Host:      "localhost",
			Name:      os.Getenv("DB_NAME"),
			Password:  os.Getenv("DB_PASSWORD"),
			Port:      "3306",
			TimeZone:  "UTC",
			TxTimeout: defaultTimeouts,
			User:      os.Getenv("DB_USER"),
		}),
		bux.WithPaymailSupport([]string{"test.com"}, "example@test.com", false, false),
		bux.WithAutoMigrate(bux.BaseModels...),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
