package chainstate

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/mrz1836/go-whatsonchain"
)

// MonitorHandler interface
type MonitorHandler interface {
	whatsonchain.SocketHandler
	SetMonitor(monitor *Monitor)
	RecordTransaction(ctx context.Context, txHex string) error
	RecordBlockHeader(ctx context.Context, bh bc.BlockHeader) error
	GetWhatsOnChain() whatsonchain.ClientInterface
}
