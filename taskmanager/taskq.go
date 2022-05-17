package taskmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis_rate/v9"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"github.com/vmihailenco/taskq/v3/redisq"
)

var (
	mutex sync.Mutex
)

// DefaultTaskQConfig will return a default configuration that can be modified
func DefaultTaskQConfig(name string) *taskq.QueueOptions {
	return &taskq.QueueOptions{
		BufferSize:           10,                      // Size of the buffer where reserved messages are stored.
		ConsumerIdleTimeout:  6 * time.Hour,           // ConsumerIdleTimeout Time after which the consumer need to be deleted.
		Handler:              nil,                     // Optional message handler. The default is the global Tasks registry.
		MaxNumFetcher:        0,                       // Maximum number of goroutines fetching messages.
		MaxNumWorker:         10,                      // Maximum number of goroutines processing messages.
		MinNumWorker:         1,                       // Minimum number of goroutines processing messages.
		Name:                 name,                    // Queue name.
		PauseErrorsThreshold: 100,                     // Number of consecutive failures after which queue processing is paused.
		RateLimit:            redis_rate.Limit{},      // Processing rate limit.
		RateLimiter:          nil,                     // Optional rate limiter. The default is to use Redis.
		Redis:                nil,                     // Redis client that is used for storing metadata.
		ReservationSize:      10,                      // Number of messages reserved by a fetcher in the queue in one request.
		ReservationTimeout:   60 * time.Second,        // Time after which the reserved message is returned to the queue.
		Storage:              taskq.NewLocalStorage(), // Optional storage interface. The default is to use Redis.
		WaitTimeout:          3 * time.Second,         // Time that a long polling receive call waits for a message to become available before returning an empty response.
		WorkerLimit:          0,                       // Global limit of concurrently running workers across all servers. Overrides MaxNumWorker.
	}
}

// convertTaskToTaskQ will convert our internal task to a TaskQ struct
func convertTaskToTaskQ(task *Task) *taskq.TaskOptions {
	return &taskq.TaskOptions{
		Name:            task.Name,
		Handler:         task.Handler,
		FallbackHandler: task.FallbackHandler,
		DeferFunc:       task.DeferFunc,
		RetryLimit:      task.RetryLimit,
		MinBackoff:      task.MinBackoff,
		MaxBackoff:      task.MaxBackoff,
	}
}

// loadTaskQ will load TaskQ based on the Factory Type and configuration set by the client loading
func (c *Client) loadTaskQ() error {

	// Check for a valid config (set on client creation)
	if c.options.taskq.config == nil {
		return ErrMissingTaskQConfig
	}

	// Load using in-memory vs Redis
	if c.options.taskq.factoryType == FactoryMemory {

		// Create the factory
		c.options.taskq.factory = memqueue.NewFactory()

	} else if c.options.taskq.factoryType == FactoryRedis {

		// Check for a redis connection (given on taskq configuration)
		if c.options.taskq.config.Redis == nil {
			return ErrMissingRedis
		}

		// Create the factory
		c.options.taskq.factory = redisq.NewFactory()

	} else {
		return ErrMissingFactory
	}

	// Set the queue
	c.options.taskq.queue = c.options.taskq.factory.RegisterQueue(c.options.taskq.config)

	// turn off logger for now
	// NOTE: having issues with logger with system resources
	// taskq.SetLogger(nil)

	return nil
}

// registerTaskUsingTaskQ will register a new task using the TaskQ engine
func (c *Client) registerTaskUsingTaskQ(task *Task) {

	defer func() {
		if err := recover(); err != nil {
			c.DebugLog(fmt.Sprintf("registering task panic: %v", err))
		}
	}()

	mutex.Lock()

	// Check if task is already registered
	if t := taskq.Tasks.Get(task.Name); t != nil {

		// Register the task locally
		c.options.taskq.tasks[task.Name] = t

		// Task already exists!
		// c.DebugLog(fmt.Sprintf("registering task: %s... task already exists!", task.Name))

		mutex.Unlock()

		return
	}

	// Register and store the task
	c.options.taskq.tasks[task.Name] = taskq.RegisterTask(convertTaskToTaskQ(task))

	mutex.Unlock()

	// Debugging
	c.DebugLog(fmt.Sprintf("registering task: %s...", c.options.taskq.tasks[task.Name].Name()))
}

// runTaskUsingTaskQ will run a task using TaskQ
func (c *Client) runTaskUsingTaskQ(ctx context.Context, options *TaskOptions) error {

	// Starting the execution of the task
	c.DebugLog(fmt.Sprintf(
		"executing task: %s... delay: %s arguments: %s",
		options.TaskName,
		options.Delay.String(),
		fmt.Sprintf("%+v", options.Arguments),
	))

	// Try to get the task
	if _, ok := c.options.taskq.tasks[options.TaskName]; !ok {
		return ErrTaskNotFound
	}

	// Add arguments, and delay if set
	msg := c.options.taskq.tasks[options.TaskName].WithArgs(ctx, options.Arguments...)
	if options.OnceInPeriod > 0 {
		msg.OnceInPeriod(options.OnceInPeriod, options.Arguments...)
	} else if options.Delay > 0 {
		msg.SetDelay(options.Delay)
	}

	// This is the "cron" aspect of the task
	if options.RunEveryPeriod > 0 {
		_, err := c.options.cronService.AddFunc(
			fmt.Sprintf("@every %ds", int(options.RunEveryPeriod.Seconds())),
			func() {
				// todo: log the error if it occurs? Cannot pass the error back up
				_ = c.options.taskq.queue.Add(msg)
			},
		)
		return err
	}

	// Add to the queue
	return c.options.taskq.queue.Add(msg)
}
