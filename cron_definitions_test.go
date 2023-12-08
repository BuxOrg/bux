package bux

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCronTasks(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		ctx := context.Background()

		clientInterface, err := NewClient(ctx, DefaultClientOpts(true, true)...)
		require.NoError(t, err)

		client := clientInterface.(*Client)

		times := make(chan time.Time)
		counter := 0
		desiredExecutions := 0
		desiredPeriod := 100 * time.Millisecond

		err = registerCronTasks(client, []cronDefinition{
			{
				Name:   "test",
				Period: desiredPeriod,
				Handler: func(ctx context.Context, client *Client) error {
					if counter < desiredExecutions {
						counter++
						times <- time.Now()
					} else {
						close(times)
					}
					return nil
				},
			},
		})

		require.NoError(t, err)

		// collect execution times from channel until it's closed after desiredExecutions
		executionTimes := []time.Time{}
		for range times {
			executionTimes = append(executionTimes, <-times)
		}

		// check number of executions
		require.Equal(t, desiredExecutions, len(executionTimes))

		// check time difference between executions
		for i := 1; i < len(executionTimes); i++ {
			diff := executionTimes[i].Sub(executionTimes[i-1])
			require.True(t, diff >= desiredPeriod && diff <= 2*desiredPeriod)
		}
	})
}
