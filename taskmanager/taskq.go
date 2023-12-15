package taskmanager

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	taskq "github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"github.com/vmihailenco/taskq/v3/redisq"
)

var mutex sync.Mutex

// TasqOps allow functional options to be supplied
type TasqOps func(*taskq.QueueOptions)

// WithRedis will set the redis client for the TaskQ engine
func WithRedis(addr string) TasqOps {
	return func(queueOptions *taskq.QueueOptions) {
		queueOptions.Redis = redis.NewClient(&redis.Options{
			Addr: strings.Replace(addr, "redis://", "", -1),
		})
	}
}

// DefaultTaskQConfig will return a QueueOptions with specified name and functional options applied
func DefaultTaskQConfig(name string, opts ...TasqOps) *taskq.QueueOptions {
	queueOptions := &taskq.QueueOptions{
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

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(queueOptions)
	}

	return queueOptions
}

// TaskRunOptions are the options for running a task
type TaskRunOptions struct {
	Arguments      []interface{} // Arguments for the task
	Delay          time.Duration // Run after X delay
	OnceInPeriod   time.Duration // Run once in X period
	RunEveryPeriod time.Duration // Cron job!
	TaskName       string        // Name of the task
}

// loadTaskQ will load TaskQ based on the Factory Type and configuration set by the client loading
func (c *TaskManager) loadTaskQ() error {
	// Check for a valid config (set on client creation)
	factoryType := c.Factory()
	if factoryType == FactoryEmpty {
		return fmt.Errorf("missing factory type to load taskq")
	}

	var factory taskq.Factory
	if factoryType == FactoryMemory {
		factory = memqueue.NewFactory()
	} else if factoryType == FactoryRedis {
		factory = redisq.NewFactory()
	}

	// Set the queue
	c.options.taskq.queue = factory.RegisterQueue(c.options.taskq.config)

	// turn off logger for now
	// NOTE: having issues with logger with system resources
	// taskq.SetLogger(nil)

	return nil
}

// RegisterTask will register a new task using the TaskQ engine
func (c *TaskManager) RegisterTask(name string, handler interface{}) (err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf(fmt.Sprintf("registering task panic: %v", panicErr))
		}
	}()

	mutex.Lock()
	defer mutex.Unlock()

	if t := taskq.Tasks.Get(name); t != nil {
		// if already registered - register the task locally
		c.options.taskq.tasks[name] = t
	} else {
		// Register and store the task
		c.options.taskq.tasks[name] = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       name,
			Handler:    handler,
			RetryLimit: 1,
		})
	}

	// Debugging
	c.DebugLog(fmt.Sprintf("registering task: %s...", c.options.taskq.tasks[name].Name()))
	return nil
}

// RunTask will run a task using TaskQ
func (c *TaskManager) RunTask(ctx context.Context, options *TaskRunOptions) error {
	// Starting the execution of the task
	c.DebugLog(fmt.Sprintf(
		"executing task: %s... delay: %s arguments: %s",
		options.TaskName,
		options.Delay.String(),
		fmt.Sprintf("%+v", options.Arguments),
	))

	// Try to get the task
	if _, ok := c.options.taskq.tasks[options.TaskName]; !ok {
		return fmt.Errorf("task %s not registered", options.TaskName)
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
