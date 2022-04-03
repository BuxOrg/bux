package chainstate

import (
	"context"

	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
)

// Monitor starts a new monitorConfig to monitorConfig and filter transactions from a source
type Monitor struct {
	logger                  Logger
	client                  *centrifuge.Client
	processor               MonitorProcessor
	centrifugeServer        string
	monitorDays             int
	falsePositiveRate       float64
	maxNumberOfDestinations int
}

// MonitorOptions options for starting this monitorConfig
type MonitorOptions struct {
	CentrifugeServer        string
	MonitorDays             int
	FalsePositiveRate       float64
	MaxNumberOfDestinations int
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
func NewMonitor(ctx context.Context, options *MonitorOptions) *Monitor {
	options.checkDefaults()
	monitor := &Monitor{
		centrifugeServer:        options.CentrifugeServer,
		maxNumberOfDestinations: options.MaxNumberOfDestinations,
		falsePositiveRate:       options.FalsePositiveRate,
		monitorDays:             options.MonitorDays,
	}
	monitor.processor = NewBloomProcessor(uint(monitor.maxNumberOfDestinations), monitor.falsePositiveRate)

	// Set logger if not set
	if monitor.logger == nil {
		monitor.logger = newLogger()
	}

	return monitor
}

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

// Monitor open a socket to the service provider and monitorConfig transactions
func (m *Monitor) Monitor(handler MonitorHandler) error {

	if m.client == nil {
		handler.SetMonitor(m)
		m.client = newClient(m.centrifugeServer, handler)
	}

	return m.client.Connect()
}

// PauseMonitor closes the monitoring socket and pauses monitoring
func (m *Monitor) PauseMonitor() error {

	return m.client.Disconnect()
}
