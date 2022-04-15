package chainstate

import (
	"context"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/logger"
	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
)

// Monitor starts a new monitorConfig to monitorConfig and filter transactions from a source
type Monitor struct {
	debug                   bool
	chainstateOptions       *clientOptions
	logger                  logger.Interface
	client                  *centrifuge.Client
	processor               MonitorProcessor
	connected               bool
	centrifugeServer        string
	monitorDays             int
	saveDestinations        bool
	falsePositiveRate       float64
	maxNumberOfDestinations int
	processMempoolOnConnect bool
	handler                 MonitorHandler
}

// MonitorOptions options for starting this monitorConfig
type MonitorOptions struct {
	Debug                   bool
	CentrifugeServer        string
	MonitorDays             int
	FalsePositiveRate       float64
	MaxNumberOfDestinations int
	ProcessMempoolOnConnect bool
	SaveDestinations        bool
}

func (o *MonitorOptions) checkDefaults() {
	if o.MonitorDays == 0 {
		o.MonitorDays = 7
	}
	if o.FalsePositiveRate == 0 {
		o.FalsePositiveRate = 0.01
	}
	if o.MaxNumberOfDestinations == 0 {
		o.MaxNumberOfDestinations = 100000
	}
}

func newClient(wsURL string, handler whatsonchain.SocketHandler) *centrifuge.Client {
	c := centrifuge.NewJsonClient(wsURL, centrifuge.DefaultConfig())

	c.OnConnect(handler)
	c.OnDisconnect(handler)
	c.OnMessage(handler)
	c.OnError(handler)

	c.OnServerPublish(handler)
	c.OnServerSubscribe(handler)
	c.OnServerUnsubscribe(handler)
	c.OnServerJoin(handler)
	c.OnServerLeave(handler)

	return c
}

// NewMonitor starts a new monitorConfig and loads all addresses that need to be monitored into the bloom filter
func NewMonitor(_ context.Context, options *MonitorOptions) *Monitor {
	options.checkDefaults()
	monitor := &Monitor{
		debug:                   options.Debug,
		centrifugeServer:        options.CentrifugeServer,
		maxNumberOfDestinations: options.MaxNumberOfDestinations,
		falsePositiveRate:       options.FalsePositiveRate,
		monitorDays:             options.MonitorDays,
		saveDestinations:        options.SaveDestinations,
		processMempoolOnConnect: options.ProcessMempoolOnConnect,
	}
	// Set logger if not set
	if monitor.logger == nil {
		monitor.logger = logger.NewLogger(true)
	}

	monitor.processor = NewBloomProcessor(uint(monitor.maxNumberOfDestinations), monitor.falsePositiveRate)
	monitor.processor.Debug(options.Debug)
	monitor.processor.SetLogger(monitor.logger)

	return monitor
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

// SaveDestinations gets whether or not we should save destinations from transactions that pass monitor filter
func (m *Monitor) SaveDestinations() bool {
	return m.saveDestinations
}

// SetChainstateOptions sets the chainstate options on the monitor to allow more synching capabilities
func (m *Monitor) SetChainstateOptions(options *clientOptions) {
	m.chainstateOptions = options
}

// Monitor open a socket to the service provider and monitorConfig transactions
func (m *Monitor) Monitor(handler MonitorHandler) error {

	if m.client == nil {
		handler.SetMonitor(m)
		m.handler = handler
		m.logger.Info(context.Background(), "[MONITOR] Connecting to server: %s", m.centrifugeServer)
		m.client = newClient(m.centrifugeServer, handler)
	}

	return m.client.Connect()
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

// PauseMonitor closes the monitoring socket and pauses monitoring
func (m *Monitor) PauseMonitor() error {
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
						txHex, err = m.processor.FilterMempoolTx(tx.Hex) // todo off
						if err != nil {
							m.logger.Error(ctx, fmt.Sprintf("[MONITOR] ERROR filtering tx %s: %s\n", tx.TxID, err.Error()))
							continue
						}
						if txHex != "" {
							err = m.handler.RecordTransaction(ctx, txHex)
							if err != nil {
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
		}()
	}

	return nil
}
