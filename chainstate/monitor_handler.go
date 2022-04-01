package chainstate

import (
	"context"

	"github.com/mrz1836/go-whatsonchain"
)

// MonitorHandler interface
type MonitorHandler interface {
	whatsonchain.SocketHandler
	SetMonitor(monitor *Monitor)
	RecordTransaction(ctx context.Context, xPubKey, txHex, draftID string) error
	GetWhatsOnChain() whatsonchain.ClientInterface
}
