package chainstate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BuxOrg/bux/logger"
	"github.com/mrz1836/go-whatsonchain"
)

// Monitor starts a new monitorConfig to monitor and filter transactions from a source
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
	logger                       logger.Interface
	maxNumberOfDestinations      int
	mempoolSyncChannel           chan bool
	monitorDays                  int
	processMempoolOnConnect      bool
	processor                    MonitorProcessor
	saveTransactionsDestinations bool
}

// MonitorOptions options for starting this monitorConfig
type MonitorOptions struct {
	AuthToken                   string  `json:"token"`
	BuxAgentURL                 string  `json:"bux_agent_url"`
	Debug                       bool    `json:"debug"`
	FalsePositiveRate           float64 `json:"false_positive_rate"`
	LoadMonitoredDestinations   bool    `json:"load_monitored_destinations"`
	MaxNumberOfDestinations     int     `json:"max_number_of_destinations"`
	MonitorDays                 int     `json:"monitor_days"`
	ProcessMempoolOnConnect     bool    `json:"process_mempool_on_connect"`
	ProcessorType               string  `json:"processor_type"`
	SaveTransactionDestinations bool    `json:"save_transaction_destinations"`
}

func (o *MonitorOptions) checkDefaults() {
	if o.MonitorDays <= 0 {
		o.MonitorDays = defaultMonitorDays
	}
	if o.FalsePositiveRate <= 0 {
		o.FalsePositiveRate = defaultFalsePositiveRate
	}
	if o.MaxNumberOfDestinations <= 0 {
		o.MaxNumberOfDestinations = defaultMaxNumberOfDestinations
	}
}

// NewMonitor starts a new monitorConfig and loads all addresses that need to be monitored into the bloom filter
func NewMonitor(_ context.Context, options *MonitorOptions) (monitor *Monitor) {
	options.checkDefaults()
	if options.ProcessorType == "" {
		options.ProcessorType = FilterBloom
	}
	monitor = &Monitor{
		authToken:                    options.AuthToken,
		buxAgentURL:                  options.BuxAgentURL,
		debug:                        options.Debug,
		falsePositiveRate:            options.FalsePositiveRate,
		filterType:                   options.ProcessorType,
		maxNumberOfDestinations:      options.MaxNumberOfDestinations,
		mempoolSyncChannel:           make(chan bool),
		monitorDays:                  options.MonitorDays,
		processMempoolOnConnect:      options.ProcessMempoolOnConnect,
		saveTransactionsDestinations: options.SaveTransactionDestinations,
	}

	// Set logger if not set
	if monitor.logger == nil {
		monitor.logger = logger.NewLogger(options.Debug, 4)
	}

	switch monitor.filterType {
	case FilterRegex:
		monitor.processor = NewRegexProcessor()
	case FilterBloom:
	default:
		monitor.processor = NewBloomProcessor(uint(monitor.maxNumberOfDestinations), monitor.falsePositiveRate)
	}

	monitor.processor.Debug(options.Debug)
	monitor.processor.SetLogger(monitor.logger)
	return
}

// IsDebug gets whether debugging is on
func (m *Monitor) IsDebug() bool {
	return m.debug
}

// Logger gets the current logger
func (m *Monitor) Logger() Logger {
	return m.logger
}

// Processor gets the monitor processor
func (m *Monitor) Processor() MonitorProcessor {
	return m.processor
}

// GetMonitorDays gets the monitorDays option
func (m *Monitor) GetMonitorDays() int {
	return m.monitorDays
}

// GetFalsePositiveRate gets the falsePositiveRate option
func (m *Monitor) GetFalsePositiveRate() float64 {
	return m.falsePositiveRate
}

// GetMaxNumberOfDestinations gets the monitorDays option
func (m *Monitor) GetMaxNumberOfDestinations() int {
	return m.maxNumberOfDestinations
}

// GetProcessMempoolOnConnect gets whether the whole mempool should be processed when connecting
func (m *Monitor) GetProcessMempoolOnConnect() bool {
	return m.processMempoolOnConnect
}

// SaveDestinations gets whether we should save destinations from transactions that pass monitor filter
func (m *Monitor) SaveDestinations() bool {
	return m.saveTransactionsDestinations
}

// LoadMonitoredDestinations gets where we want to add the monitored destinations from the database into the processor
func (m *Monitor) LoadMonitoredDestinations() bool {
	return m.loadMonitoredDestinations
}

// SetChainstateOptions sets the chainstate options on the monitor to allow more syncing capabilities
func (m *Monitor) SetChainstateOptions(options *clientOptions) {
	m.chainstateOptions = options
}

// Start open a socket to the service provider and monitorConfig transactions
func (m *Monitor) Start(ctx context.Context, handler MonitorHandler) error {
	if m.client == nil {
		handler.SetMonitor(m)
		m.handler = handler
		m.logger.Info(ctx, fmt.Sprintf("[MONITOR] Starting, connecting to server: %s", m.buxAgentURL))
		m.client = newCentrifugeClient(m.buxAgentURL, handler)
		if m.authToken != "" {
			m.client.SetToken(m.authToken)
		}
	}

	return m.client.Connect()
}

// Add a new item to monitor
func (m *Monitor) Add(regexString, item string) error {
	if m.processor == nil {
		return ErrMonitorNotAvailable
	}
	// todo signal to bux-agent that a new item was added
	_, err := m.client.AddFilter(regexString, item)
	if err != nil {
		return err
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

// IsConnected returns whether we are connected to the socket
func (m *Monitor) IsConnected() bool {
	return m.connected
}

// Stop closes the monitoring socket and pauses monitoring
func (m *Monitor) Stop(ctx context.Context) error {
	m.logger.Info(ctx, "[MONITOR] Stopping monitor...")
	defer close(m.mempoolSyncChannel)
	return m.client.Disconnect()
}

// ProcessMempool processes all current transactions in the mempool
func (m *Monitor) ProcessMempool(ctx context.Context) error {

	woc := m.handler.GetWhatsOnChain()
	if woc != nil {
		mempoolTxs, err := woc.GetMempoolTransactions(ctx)
		if err != nil {
			return err
		}

		// TODO: This is overkill right now, but gives us a chance to parallelize this stuff
		var done sync.WaitGroup
		done.Add(1)

		// run the processing of the txs in a different thread
		go func() {
			if m.debug {
				m.logger.Info(ctx, fmt.Sprintf("[MONITOR] ProcessMempool mempoolTxs: %d\n", len(mempoolTxs)))
			}
			if len(mempoolTxs) > 0 {
				hashes := new(whatsonchain.TxHashes)
				hashes.TxIDs = append(hashes.TxIDs, mempoolTxs...)

				// Break up the transactions into batches
				var batches [][]string
				chunkSize := whatsonchain.MaxTransactionsRaw

				for i := 0; i < len(hashes.TxIDs); i += chunkSize {
					end := i + chunkSize
					if end > len(hashes.TxIDs) {
						end = len(hashes.TxIDs)
					}
					batches = append(batches, hashes.TxIDs[i:end])
				}
				if m.debug {
					m.logger.Info(ctx, fmt.Sprintf("[MONITOR] ProcessMempool created batches: %d\n", len(batches)))
				}

				var currentRateLimit int
				// Loop Batches - and get each batch (multiple batches of MaxTransactionsRaw)
				// this code comes from the go-whatsonchain lib, but we want to process per 20
				// and not the whole batch in 1 go
				for i, batch := range batches {
					if m.debug {
						m.logger.Info(ctx, fmt.Sprintf("[MONITOR] ProcessMempool processing batch: %d\n", i+1))
					}
					// While processing all the batches, check if channel is closed
					select {
					case <-m.mempoolSyncChannel:
						return
					default:

						txHashes := new(whatsonchain.TxHashes)
						txHashes.TxIDs = append(txHashes.TxIDs, batch...)

						// Get the tx details (max of MaxTransactionsUTXO)
						var returnedList whatsonchain.TxList
						if returnedList, err = woc.BulkRawTransactionDataProcessor(
							ctx, txHashes,
						); err != nil {
							return
						}

						// Add to the list
						for _, tx := range returnedList {
							if m.debug {
								m.logger.Info(ctx, fmt.Sprintf("[MONITOR] ProcessMempool tx: %s\n", tx.TxID))
							}
							var txHex string
							txHex, err = m.processor.FilterTransaction(tx.Hex) // todo off
							if err != nil {
								m.logger.Error(ctx, fmt.Sprintf("[MONITOR] ERROR filtering tx %s: %s\n", tx.TxID, err.Error()))
								continue
							}
							if txHex != "" {
								if err = m.handler.RecordTransaction(ctx, txHex); err != nil {
									m.logger.Error(ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %s\n", err.Error()))
									continue
								}
								if m.debug {
									m.logger.Info(ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %s\n", tx.TxID))
								}
							}
						}

						// Accumulate / sleep to prevent rate limiting
						currentRateLimit++
						if currentRateLimit >= woc.RateLimit() {
							time.Sleep(1 * time.Second)
							currentRateLimit = 0
						}
					}
				}
			}
			done.Done()
		}()
		done.Wait()
		m.mempoolSyncChannel <- true
	}

	return nil
}
