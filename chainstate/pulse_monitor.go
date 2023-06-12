package chainstate

import (
	"context"
	"fmt"

	zLogger "github.com/mrz1836/go-logger"
)

// Pulse struct
type Pulse struct {
	authToken string
	pulseURL  string
	client    PulseMonitorClient
	connected bool
	handler   PulseHandler
	logger    zLogger.GormLoggerInterface
	debug     bool
}

// PulseOptions options for starting Pulse
type PulseOptions struct {
	AuthToken string `json:"token"`
	PulseURL  string `json:"pulse_url"`
	Debug     bool   `json:"debug"`
}

// NewPulse starts a new Pulse
func NewPulse(_ context.Context, options *PulseOptions) (pulse *Pulse) {
	pulse = &Pulse{
		authToken: options.AuthToken,
		pulseURL:  options.PulseURL,
		logger:    zLogger.NewGormLogger(options.Debug, 4),
		debug:     options.Debug,
	}

	return
}

// Connected sets the connected state to true
func (p *Pulse) Connected() {
	p.connected = true
}

// Disconnected sets the connected state to false
func (p *Pulse) Disconnected() {
	p.connected = false
}

// IsConnected returns whether we are connected to the socket
func (p *Pulse) IsConnected() bool {
	return p.connected
}

// IsDebug returns the debug flag (bool)
func (p *Pulse) IsDebug() bool {
	return p.debug
}

// Start open a socket to Pulse
func (p *Pulse) Start(ctx context.Context, handler PulseHandler) error {
	if p.client == nil {
		handler.SetPulse(p)
		p.handler = handler
		p.logger.Info(ctx, fmt.Sprintf("[PULSE] Starting, connecting to server: %s", p.pulseURL))
		p.client = newPulseCentrifugeClient(p.pulseURL, p.authToken, handler)
	}

	err := p.client.Connect()
	if err != nil {
		p.logger.Error(ctx, fmt.Sprintf("[PULSE] Connection error. URL: %s, ERROR: %s", p.pulseURL, err.Error()))
		return err
	}
	err = p.client.Subscribe(handler)
	if err != nil {
		p.logger.Error(ctx, fmt.Sprintf("[PULSE Subscription] Subscription error. URL: %s, ERROR: %s", p.pulseURL, err.Error()))
		return err
	}
	return nil
}

// Stop closes the pulse socket and pauses monitoring
func (p *Pulse) Stop(ctx context.Context) error {
	p.logger.Info(ctx, "[PULSE] Stopping monitor...")
	if p.IsConnected() {
		return p.client.Disconnect()
	}

	return nil
}

// Logger gets the current logger
func (p *Pulse) Logger() Logger {
	return p.logger
}
