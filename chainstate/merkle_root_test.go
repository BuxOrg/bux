package chainstate

import (
	"bytes"
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func initMockClient(ops ...ClientOps) (*Client, *buffLogger) {
	bLogger := newBuffLogger()
	ops = append(ops, WithLogger(bLogger.logger))
	c, _ := NewClient(
		context.Background(),
		ops...,
	)
	return c.(*Client), bLogger
}

func TestVerifyMerkleRoots(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockURL := "http://pulse.test/api/v1/chain/merkleroot/verify"

	t.Run("no pulse client", func(t *testing.T) {
		c, _ := initMockClient()

		err := c.VerifyMerkleRoots(context.Background(), []MerkleRootConfirmationRequestItem{})

		assert.Error(t, err)
	})

	t.Run("pulse is not online", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder("POST", mockURL,
			httpmock.NewStringResponder(500, `{"error":"Internal Server Error"}`),
		)
		c, bLogger := initMockClient(WithConnectionToPulse(mockURL, ""))

		err := c.VerifyMerkleRoots(context.Background(), []MerkleRootConfirmationRequestItem{})

		assert.Error(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
		assert.True(t, bLogger.contains("pulse client returned status code 500"))
	})

	t.Run("pulse wrong auth", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder("POST", mockURL,
			httpmock.NewStringResponder(401, `Unauthorized`),
		)
		c, bLogger := initMockClient(WithConnectionToPulse(mockURL, "some-token"))

		err := c.VerifyMerkleRoots(context.Background(), []MerkleRootConfirmationRequestItem{})

		assert.Error(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
		assert.True(t, bLogger.contains("401"))
	})

	t.Run("pulse invalid state", func(t *testing.T) {
		httpmock.Reset()
		httpmock.RegisterResponder("POST", mockURL,
			httpmock.NewJsonResponderOrPanic(200, MerkleRootsConfirmationsResponse{
				ConfirmationState: Invalid,
				Confirmations:     []MerkleRootConfirmation{},
			}),
		)
		c, bLogger := initMockClient(WithConnectionToPulse(mockURL, "some-token"))

		err := c.VerifyMerkleRoots(context.Background(), []MerkleRootConfirmationRequestItem{})

		assert.Error(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
		assert.True(t, bLogger.contains("Not all merkle roots confirmed"))
	})
}

// buffLogger allows to check if a certain string was logged
type buffLogger struct {
	logger *zerolog.Logger
	buf    *bytes.Buffer
}

func newBuffLogger() *buffLogger {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Logger()
	return &buffLogger{logger: &logger, buf: &buf}
}

func (l *buffLogger) contains(expected string) bool {
	return bytes.Contains(l.buf.Bytes(), []byte(expected))
}
