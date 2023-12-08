package taskmanager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCronTasks(t *testing.T) {
	makeClient := func() *Client {
		client, _ := NewClient(
			context.Background(),
			WithTaskQ(DefaultTaskQConfig(testQueueName), FactoryMemory),
		)
		return client.(*Client)
	}

	t.Run("register one cron job ", func(t *testing.T) {
		client := makeClient()

		desiredExecutions := 0
		desiredPeriod := 100 * time.Millisecond

		type mockTarget struct {
			times   chan time.Time
			counter int
		}
		target := &mockTarget{make(chan time.Time), 0}

		err := client.CronJobsInit(target, []CronJob{
			{
				Name:   "test",
				Period: desiredPeriod,
				Handler: func(ctx context.Context, target interface{}) error {
					t := target.(*mockTarget)
					if t.counter < desiredExecutions {
						t.counter++
						t.times <- time.Now()
					} else {
						close(t.times)
					}
					return nil
				},
			},
		})

		require.NoError(t, err)

		// collect execution times from channel until it's closed after desiredExecutions
		executionTimes := []time.Time{}
		for range target.times {
			executionTimes = append(executionTimes, <-target.times)
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
