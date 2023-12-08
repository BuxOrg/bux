package chainstate

import (
	"context"
	"fmt"

	"github.com/BuxOrg/bux/logging"
	"github.com/BuxOrg/bux/utils"
	"github.com/rs/zerolog"
)

// Monitor starts a new monitorConfig to monitor and filter transactions from a source
//
// Internal struct with all options being private
type Monitor struct {
	authToken                    string
	buxAgentURL                  string
	chainstateOptions            *clientOptions
	client                       MonitorClient
	connected                    bool
	debug                        bool
	falsePositiveRate            float64
	filterType                   string
	handler                      MonitorHandler
	loadMonitoredDestinations    bool
	lockID                       string
	logger                       *zerolog.Logger
	maxNumberOfDestinations      int
	mempoolSyncChannelActive     bool
	mempoolSyncChannel           chan bool
	monitorDays                  int
	processor                    MonitorProcessor
	saveTransactionsDestinations bool
	onStop                       func()
	allowUnknownTransactions     bool
}

// MonitorOptions options for starting this monitorConfig
type MonitorOptions struct {
	AuthToken                   string  `json:"token"`
	BuxAgentURL                 string  `json:"bux_agent_url"`
	Debug                       bool    `json:"debug"`
	FalsePositiveRate           float64 `json:"false_positive_rate"`
	LoadMonitoredDestinations   bool    `json:"load_monitored_destinations"`
	LockID                      string  `json:"lock_id"`
	MaxNumberOfDestinations     int     `json:"max_number_of_destinations"`
	MonitorDays                 int     `json:"monitor_days"`
	ProcessorType               string  `json:"processor_type"`
	SaveTransactionDestinations bool    `json:"save_transaction_destinations"`
	AllowUnknownTransactions    bool    `json:"allow_unknown_transactions"` // whether to allow transactions that do not have an xpub_in_id or xpub_out_id
}

// checkDefaults will check for missing values and set default values
func (o *MonitorOptions) checkDefaults() {
	// Set the default for Monitor Days (days in past)
	if o.MonitorDays <= 0 {
		o.MonitorDays = defaultMonitorDays
	}

	// Set the false positive rate
	if o.FalsePositiveRate <= 0 {
		o.FalsePositiveRate = defaultFalsePositiveRate
	}

	// Set the maximum number of destinations to monitor
	if o.MaxNumberOfDestinations <= 0 {
		o.MaxNumberOfDestinations = defaultMaxNumberOfDestinations
	}

	// Set a unique lock id if it's not provided
	if len(o.LockID) == 0 { // todo: lockID should always be set (return an error if not set?)
		o.LockID, _ = utils.RandomHex(32)
	}
}

// NewMonitor starts a new monitorConfig and loads all addresses that need to be monitored into the bloom filter
func NewMonitor(_ context.Context, options *MonitorOptions) (monitor *Monitor) {
	// Check the defaults
	options.checkDefaults()

	// Set the default processor type if not recognized
	if options.ProcessorType != FilterBloom && options.ProcessorType != FilterRegex {
		options.ProcessorType = FilterBloom
	}

	// Create a monitor struct
	monitor = &Monitor{
		authToken:                    options.AuthToken,
		buxAgentURL:                  options.BuxAgentURL,
		debug:                        options.Debug,
		falsePositiveRate:            options.FalsePositiveRate,
		filterType:                   options.ProcessorType,
		loadMonitoredDestinations:    options.LoadMonitoredDestinations,
		lockID:                       options.LockID,
		maxNumberOfDestinations:      options.MaxNumberOfDestinations,
		monitorDays:                  options.MonitorDays,
		saveTransactionsDestinations: options.SaveTransactionDestinations,
		allowUnknownTransactions:     options.AllowUnknownTransactions,
	}

	// Set logger if not set
	if monitor.logger == nil {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println("monitor.logger is nil")
		monitor.logger = logging.GetDefaultLogger()
	}

	// Switch on the filter type
	switch monitor.filterType {
	case FilterRegex:
		monitor.processor = NewRegexProcessor()
	default:
		monitor.processor = NewBloomProcessor(uint(monitor.maxNumberOfDestinations), monitor.falsePositiveRate)
	}

	// Load the settings for debugging and logging
	monitor.processor.Debug(options.Debug)
	monitor.processor.SetLogger(monitor.logger)
	return
}

// Add a new item to monitor
func (m *Monitor) Add(regexString, item string) error {
	if m.processor == nil {
		return ErrMonitorNotAvailable
	}
	// todo signal to bux-agent that a new item was added
	if m.client != nil {
		if _, err := m.client.AddFilter(regexString, item); err != nil {
			return err
		}
	} else {
		m.logger.Error().Msg("client was expected but not found")
	}
	return m.processor.Add(regexString, item)
}

// Connected sets the connected state to true
func (m *Monitor) Connected() {
	m.connected = true
}

// Disconnected sets the connected state to false
func (m *Monitor) Disconnected() {
	m.connected = false
}

// GetMonitorDays gets the monitorDays option
func (m *Monitor) GetMonitorDays() int {
	return m.monitorDays
}

// GetFalsePositiveRate gets the falsePositiveRate option
func (m *Monitor) GetFalsePositiveRate() float64 {
	return m.falsePositiveRate
}

// GetLockID gets the lock id from the Monitor
func (m *Monitor) GetLockID() string {
	return m.lockID
}

// GetMaxNumberOfDestinations gets the monitorDays option
func (m *Monitor) GetMaxNumberOfDestinations() int {
	return m.maxNumberOfDestinations
}

// IsConnected returns whether we are connected to the socket
func (m *Monitor) IsConnected() bool {
	return m.connected
}

// IsDebug gets whether debugging is on
func (m *Monitor) IsDebug() bool {
	return m.debug
}

// LoadMonitoredDestinations gets where we want to add the monitored destinations from the database into the processor
func (m *Monitor) LoadMonitoredDestinations() bool {
	return m.loadMonitoredDestinations
}

// AllowUnknownTransactions gets whether we allow recording transactions with no relation to our xpubs
func (m *Monitor) AllowUnknownTransactions() bool {
	return m.allowUnknownTransactions
}

// Logger gets the current logger
func (m *Monitor) Logger() *zerolog.Logger {
	return m.logger
}

// Processor gets the monitor processor
func (m *Monitor) Processor() MonitorProcessor {
	return m.processor
}

// SaveDestinations gets whether we should save destinations from transactions that pass monitor filter
func (m *Monitor) SaveDestinations() bool {
	return m.saveTransactionsDestinations
}

// SetChainstateOptions sets the chainstate options on the monitor to allow more syncing capabilities
func (m *Monitor) SetChainstateOptions(options *clientOptions) {
	m.chainstateOptions = options
}

// Start open a socket to the service provider and monitorConfig transactions
func (m *Monitor) Start(_ context.Context, handler MonitorHandler, onStop func()) error {
	if m.client == nil {
		handler.SetMonitor(m)
		m.handler = handler
		m.logger.Info().Msgf("[MONITOR] Starting, connecting to server: %s", m.buxAgentURL)
		m.client = newCentrifugeClient(m.buxAgentURL, handler)
		if m.authToken != "" {
			m.client.SetToken(m.authToken)
		}
	}

	m.onStop = onStop

	return m.client.Connect()
}

// Stop closes the monitoring socket and pauses monitoring
func (m *Monitor) Stop(_ context.Context) error {
	m.logger.Info().Msg("[MONITOR] Stopping monitor...")
	if m.IsConnected() { // Only close if still connected
		if m.mempoolSyncChannelActive {
			close(m.mempoolSyncChannel)
			m.mempoolSyncChannelActive = false
		}
		return m.client.Disconnect()
	}

	if m.onStop != nil {
		m.onStop()
	}

	return nil
}
