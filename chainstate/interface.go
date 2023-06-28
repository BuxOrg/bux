package chainstate

import (
	"context"
	"net/http"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/go-bc"
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
	Broadcast(ctx context.Context, id, txHex string, timeout time.Duration) (string, error)
	QueryTransaction(
		ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
	) (*TransactionInfo, error)
	QueryTransactionFastest(
		ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
	) (*TransactionInfo, error)
}

// ProviderServices is the chainstate providers interface
type ProviderServices interface {
	Minercraft() minercraft.ClientInterface
	NowNodes() nownodes.ClientInterface
	WhatsOnChain() whatsonchain.ClientInterface
}

// MinercraftServices is the minercraft services interface
type MinercraftServices interface {
	BroadcastMiners() []*Miner
	QueryMiners() []*Miner
	RefreshFeeQuotes(ctx context.Context) error
	FeeUnit() *utils.FeeUnit
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
	GetWhatsOnChain() whatsonchain.ClientInterface
	RecordBlockHeader(ctx context.Context, bh bc.BlockHeader) error
	RecordTransaction(ctx context.Context, txHex string) error
	SetMonitor(monitor *Monitor)
}

// MonitorProcessor struct that defines interface to all filter processors
type MonitorProcessor interface {
	Add(regexString, item string) error
	Debug(bool)
	FilterTransaction(txHex string) (string, error)
	FilterTransactionPublishEvent(eData []byte) (string, error)
	GetFilters() map[string]*BloomProcessorFilter
	GetHash() string
	IsDebug() bool
	Logger() Logger
	Reload(regexString string, items []string) error
	SetFilter(regex string, filter []byte) error
	SetLogger(logger Logger)
	Test(regexString string, item string) bool
}

// MonitorService for the monitoring
type MonitorService interface {
	Add(regexpString string, item string) error
	Connected()
	Disconnected()
	GetFalsePositiveRate() float64
	GetLockID() string
	GetMaxNumberOfDestinations() int
	GetMonitorDays() int
	GetProcessMempoolOnConnect() bool
	IsConnected() bool
	IsDebug() bool
	LoadMonitoredDestinations() bool
	AllowUnknownTransactions() bool
	Logger() Logger
	ProcessMempool(ctx context.Context) error
	Processor() MonitorProcessor
	SaveDestinations() bool
	Start(ctx context.Context, handler MonitorHandler, onStop func()) error
	Stop(ctx context.Context) error
}
