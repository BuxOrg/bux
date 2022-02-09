package chainstate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewTestClient returns a test client
func NewTestClient(ctx context.Context, t *testing.T, opts ...ClientOps) ClientInterface {
	c, err := NewClient(
		ctx, append(opts, WithDebugging())...,
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}
