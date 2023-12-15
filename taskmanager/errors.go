package taskmanager

import "errors"

// ErrMissingTaskQConfig is when the taskq configuration is missing prior to loading taskq
var ErrMissingTaskQConfig = errors.New("missing taskq configuration")

// ErrMissingRedis is when the Redis connection is missing prior to loading taskq
var ErrMissingRedis = errors.New("missing redis connection")

// ErrMissingFactory is when the factory type is missing or empty
var ErrMissingFactory = errors.New("missing factory type to load taskq")

// ErrTaskNotFound is when a task was not found
var ErrTaskNotFound = errors.New("task not found")
