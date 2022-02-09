package taskmanager

import "errors"

// ErrNoEngine is returned when there is no engine set (missing engine)
var ErrNoEngine = errors.New("task manager engine is empty: choose taskq or machinery (IE: WithTaskQ())")

// ErrMissingTaskQConfig is when the taskq configuration is missing prior to loading taskq
var ErrMissingTaskQConfig = errors.New("missing taskq configuration")

// ErrMissingRedis is when the Redis connection is missing prior to loading taskq
var ErrMissingRedis = errors.New("missing redis connection")

// ErrMissingFactory is when the factory type is missing or empty
var ErrMissingFactory = errors.New("missing factory type to load taskq")

// ErrEngineNotSupported is when a feature is not supported by another engine
var ErrEngineNotSupported = errors.New("engine not supported")

// ErrTaskNotFound is when a task was not found
var ErrTaskNotFound = errors.New("task not found")

// ErrMissingTaskName is when the task name is missing
var ErrMissingTaskName = errors.New("missing task name")

// ErrInvalidTaskDuration is when the task duration is invalid
var ErrInvalidTaskDuration = errors.New("invalid duration for task")

// ErrNoTasksFound is when there are no tasks found in the taskmanager
var ErrNoTasksFound = errors.New("no tasks found")
