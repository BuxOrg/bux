package chainstate

import (
	"context"
	"net/http"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/go-bc"
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
	Error(ctx context.Context, message string, params ...interface{})
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
}

// ProviderServices is the chainstate providers interface
type ProviderServices interface {
	MatterCloud() mattercloud.ClientInterface
	Minercraft() minercraft.ClientInterface
	NowNodes() nownodes.ClientInterface
	WhatsOnChain() whatsonchain.ClientInterface
}

// MinercraftServices is the minercraft services interface
type MinercraftServices interface {
	BroadcastMiners() []*Miner
	Miners() []*Miner
	QueryMiners() []*Miner
	RefreshFeeQuotes(ctx context.Context) error
}

// ClientInterface is the chainstate client interface
type ClientInterface interface {
	ChainService
	ProviderServices
	MinercraftServices
	Close(ctx context.Context)
	Debug(on bool)
	DebugLog(text string)
	HTTPClient() HTTPInterface
	IsDebug() bool
	IsNewRelicEnabled() bool
	Monitor() MonitorService
	Network() Network
	QueryTimeout() time.Duration
}

// MonitorClient interface
type MonitorClient interface {
	AddFilter(regex, item string) (centrifuge.PublishResult, error)
	Connect() error
	Disconnect() error
	SetToken(token string)
}

// MonitorHandler interface
type MonitorHandler interface {
	whatsonchain.SocketHandler
	SetMonitor(monitor *Monitor)
	RecordTransaction(ctx context.Context, txHex string) error
	RecordBlockHeader(ctx context.Context, bh bc.BlockHeader) error
	GetWhatsOnChain() whatsonchain.ClientInterface
}

// MonitorProcessor struct that defines interface to all filter processors
type MonitorProcessor interface {
	Add(regexString, item string) error
	Debug(bool)
	FilterTransactionPublishEvent(eData []byte) (string, error)
	FilterTransaction(txHex string) (string, error)
	GetFilters() map[string]*BloomProcessorFilter
	SetFilter(regex string, filter []byte) error
	GetHash() string
	IsDebug() bool
	Logger() Logger
	Reload(regexString string, items []string) error
	SetLogger(logger Logger)
	Test(regexString string, item string) bool
}

// MonitorService for the monitoring
type MonitorService interface {
	IsDebug() bool
	Logger() Logger
	Connected()
	Disconnected()
	IsConnected() bool
	GetMonitorDays() int
	SaveDestinations() bool
	LoadMonitoredDestinations() bool
	GetFalsePositiveRate() float64
	GetMaxNumberOfDestinations() int
	GetProcessMempoolOnConnect() bool
	Add(regexpString string, item string) error
	Processor() MonitorProcessor
	ProcessMempool(ctx context.Context) error
	Monitor(handler MonitorHandler) error
	PauseMonitor() error
}
