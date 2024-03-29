package main

import (
	"context"
	"log"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/mrz1836/go-cachestore"
)

func main() {
	redisURL := "localhost:6379"
	client, err := bux.NewClient(
		context.Background(), // Set context
		bux.WithRedis(&cachestore.RedisConfig{URL: redisURL}), // Cache
		bux.WithTaskqConfig( // Tasks
			taskmanager.DefaultTaskQConfig("example_queue", taskmanager.WithRedis(redisURL)),
		),
	)
	if err != nil {
		log.Fatalln("error: " + err.Error())
	}

	defer func() {
		_ = client.Close(context.Background())
	}()

	log.Println("client loaded!", client.UserAgent())
}
