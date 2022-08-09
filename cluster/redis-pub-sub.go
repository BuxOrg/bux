package cluster

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisPubSub struct
type RedisPubSub struct {
	ctx           context.Context
	client        *redis.Client
	debug         bool
	options       *redis.Options
	prefix        string
	subscriptions map[string]*redis.PubSub
}

// NewRedisPubSub create a new redis pub/sub client
func NewRedisPubSub(ctx context.Context, options *redis.Options) (*RedisPubSub, error) {
	client := redis.NewClient(options)

	return &RedisPubSub{
		ctx:           ctx,
		client:        client,
		options:       options,
		subscriptions: make(map[string]*redis.PubSub),
	}, nil
}

// Subscribe to a channel
func (r *RedisPubSub) Subscribe(channel Channel, callback func(data string)) (func() error, error) {

	channelName := r.prefix + string(channel)

	if r.debug {
		fmt.Printf("NEW SUBSCRIPTION: %s -> %s\n", channel, channelName)
	}
	sub := r.client.Subscribe(r.ctx, channelName)

	go func(ch <-chan *redis.Message) {
		if r.debug {
			fmt.Printf("START CHANNEL LISTENER: %s\n", channelName)
		}
		for msg := range ch {
			if r.debug {
				fmt.Printf("NEW PUBLISH MESSAGE: %s -> %v\n", channelName, msg)
			}
			callback(msg.Payload)
		}
	}(sub.Channel())

	return func() error {
		if r.debug {
			fmt.Printf("CLOSE PUBLICATION: %s\n", channelName)
		}
		return sub.Close()
	}, nil
}

// Publish to a channel
func (r *RedisPubSub) Publish(channel Channel, data string) error {

	channelName := r.prefix + string(channel)
	if r.debug {
		fmt.Printf("PUBLISH: %s -> %s\n", channelName, data)
	}
	err := r.client.Publish(r.ctx, channelName, data)
	if err != nil {
		return err.Err()
	}

	return nil
}
