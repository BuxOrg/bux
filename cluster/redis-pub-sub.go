package cluster

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	zLogger "github.com/mrz1836/go-logger"
)

// RedisPubSub struct
type RedisPubSub struct {
	ctx           context.Context
	client        *redis.Client
	debug         bool
	logger        zLogger.GormLoggerInterface
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

// Logger returns the logger to use
func (r *RedisPubSub) Logger() zLogger.GormLoggerInterface {
	return r.logger
}

// Subscribe to a channel
func (r *RedisPubSub) Subscribe(channel Channel, callback func(data string)) (func() error, error) {

	ctx := context.Background()
	channelName := r.prefix + string(channel)

	if r.debug {
		r.Logger().Info(ctx, fmt.Sprintf("NEW SUBSCRIPTION: %s -> %s", channel, channelName))
	}
	sub := r.client.Subscribe(r.ctx, channelName)

	go func(ch <-chan *redis.Message) {
		if r.debug {
			r.Logger().Info(ctx, fmt.Sprintf("START CHANNEL LISTENER: %s", channelName))
		}
		for msg := range ch {
			if r.debug {
				r.Logger().Info(ctx, fmt.Sprintf("NEW PUBLISH MESSAGE: %s -> %v", channelName, msg))
			}
			callback(msg.Payload)
		}
	}(sub.Channel())

	return func() error {
		if r.debug {
			r.Logger().Info(ctx, fmt.Sprintf("CLOSE PUBLICATION: %s", channelName))
		}
		return sub.Close()
	}, nil
}

// Publish to a channel
func (r *RedisPubSub) Publish(channel Channel, data string) error {

	channelName := r.prefix + string(channel)
	if r.debug {
		r.Logger().Info(context.Background(), fmt.Sprintf("PUBLISH: %s -> %s", channelName, data))
	}
	err := r.client.Publish(r.ctx, channelName, data)
	if err != nil {
		return err.Err()
	}

	return nil
}
