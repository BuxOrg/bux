package chainstate

import (
	"context"
	"net/http"
	"time"

	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/tonicpow/go-minercraft"
)

// HTTPInterface is the HTTP client interface
type HTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// Logger is the logger interface for debug messages
type Logger interface {
	Info(ctx context.Context, message string, params ...interface{})
}

// ChainService is the chain related methods
type ChainService interface {
	Broadcast(ctx context.Context, id, txHex string, timeout time.Duration) error
	QueryTransaction(
		ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
	) (*TransactionInfo, error)
	QueryTransactionFastest(
		ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
	) (*TransactionInfo, error)
	MonitorMempool(ctx context.Context, filter string) error
	MonitorBlockHeaders(ctx context.Context) error
}

// ProviderServices is the chainstate providers interface
type ProviderServices interface {
	MatterCloud() mattercloud.ClientInterface
	Minercraft() minercraft.ClientInterface
	NowNodes() nownodes.ClientInterface
	WhatsOnChain() whatsonchain.ClientInterface
}

// ClientInterface is the chainstate client interface
type ClientInterface interface {
	ChainService
	ProviderServices
	BroadcastMiners() []*minercraft.Miner
	Close(ctx context.Context)
	Debug(on bool)
	DebugLog(text string)
	HTTPClient() HTTPInterface
	IsDebug() bool
	IsNewRelicEnabled() bool
	Miners() []*minercraft.Miner
	Monitor() MonitorService
	Network() Network
	QueryMiners() []*minercraft.Miner
	QueryTimeout() time.Duration
}

// MonitorService for the monitoring
type MonitorService interface {
	GetMonitorDays() int
	GetFalsePositiveRate() float64
	GetMaxNumberOfDestinations() int
	Add(item string)
	Test(item string) bool
	Reload(items []string) error
	Monitor() error
	PauseMonitor() error
}
