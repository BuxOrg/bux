package taskmanager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testQueueName = "test_queue"
)

// todo: finish unit tests!
// NOTE: these are just tests to get the library built ;)

func TestNewClient(t *testing.T) {
	c, err := NewClient(
		context.Background(),
		WithTaskqConfig(DefaultTaskQConfig(testQueueName, nil)),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	defer func() {
		_ = c.Close(context.Background())
	}()

	ctx := c.GetTxnCtx(context.Background())

	err = c.RegisterTask(&Task{
		Name: "task-1",
		Handler: func(name string) error {
			fmt.Println("TSK1 ran: " + name)
			return nil
		},
	})
	require.NoError(t, err)

	err = c.RegisterTask(&Task{
		Name: "task-2",
		Handler: func(name string) error {
			fmt.Println("TSK2 ran: " + name)
			return nil
		},
	})
	require.NoError(t, err)

	t.Log("tasks: ", len(c.Tasks()))

	time.Sleep(2 * time.Second)

	// Run tasks
	err = c.RunTask(ctx, &TaskOptions{
		Arguments: []interface{}{"task #1"},
		TaskName:  "task-1",
	})
	require.NoError(t, err)

	err = c.RunTask(ctx, &TaskOptions{
		Arguments: []interface{}{"task #2"},
		TaskName:  "task-2",
	})
	require.NoError(t, err)

	err = c.RunTask(ctx, &TaskOptions{
		Arguments: []interface{}{"task #2 with delay"},
		Delay:     time.Second,
		TaskName:  "task-2",
	})
	require.NoError(t, err)

	t.Log("ran all tasks...")

	t.Log("closing...")
}
