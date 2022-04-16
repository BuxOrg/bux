package bux

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/libsv/go-bc"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
	"github.com/korovkin/limiter"
	"github.com/mrz1836/go-whatsonchain"
)

// MonitorEventHandler for handling transaction events from a monitor
type MonitorEventHandler struct {
	debug     bool
	logger    chainstate.Logger
	monitor   chainstate.MonitorService
	buxClient ClientInterface
	ctx       context.Context
	limit     *limiter.ConcurrencyLimiter
}

// NewMonitorHandler create a new monitor handler
func NewMonitorHandler(ctx context.Context, buxClient ClientInterface, monitor chainstate.MonitorService) MonitorEventHandler {
	return MonitorEventHandler{
		debug:     monitor.IsDebug(),
		logger:    monitor.Logger(),
		monitor:   monitor,
		buxClient: buxClient,
		ctx:       ctx,
		limit:     limiter.NewConcurrencyLimiter(runtime.NumCPU()),
	}
}

// OnConnect event when connected
func (h *MonitorEventHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Connected to server: %s", e.ClientID))
	if h.monitor.GetProcessMempoolOnConnect() {
		h.logger.Info(h.ctx, "[MONITOR] PROCESS MEMPOOL")
		go func() {
			err := h.monitor.ProcessMempool(h.ctx)
			if err != nil {
				h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR processing mempool: %s", err.Error()))
			}
		}()
	}
	h.monitor.Connected()
}

// OnError on error event
func (h *MonitorEventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] Error: %s", e.Message))
}

// OnMessage on new message event
func (h *MonitorEventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	var data map[string]interface{}
	err := json.Unmarshal(e.Data, &data)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] failed unmarshalling data: %s", err.Error()))
	}

	if _, ok := data["time"]; !ok {
		if h.debug {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnMessage: %v", data))
		}
	}
}

// OnDisconnect when disconnected
func (h *MonitorEventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	h.monitor.Disconnected()
}

// OnJoin event when joining a server
func (h *MonitorEventHandler) OnJoin(_ *centrifuge.Subscription, e centrifuge.JoinEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnJoin: %v", e))
	}
}

// OnLeave event when leaving a server
func (h *MonitorEventHandler) OnLeave(_ *centrifuge.Subscription, e centrifuge.LeaveEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnLeave: %v", e))
	}
}

// OnPublish ???
func (h *MonitorEventHandler) OnPublish(_ *centrifuge.Subscription, e centrifuge.PublishEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnPublish: %v", e))
	}
}

// OnServerSubscribe ???
func (h *MonitorEventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnServerSubscribe: %v", e))
	}
}

// OnServerUnsubscribe ???
func (h *MonitorEventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnServerUnsubscribe: %v", e))
	}
}

// OnSubscribeSuccess ???
func (h *MonitorEventHandler) OnSubscribeSuccess(_ *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnSubscribeSuccess: %v", e))
	}
}

// OnSubscribeError ???
func (h *MonitorEventHandler) OnSubscribeError(_ *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnSubscribeError: %v", e))
	}
}

// OnUnsubscribe ???
func (h *MonitorEventHandler) OnUnsubscribe(_ *centrifuge.Subscription, e centrifuge.UnsubscribeEvent) {
	if h.debug {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnUnsubscribe: %v", e))
	}
}

// OnServerJoin event when joining a server
func (h *MonitorEventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Joined server: %v", e))
}

// OnServerLeave event when leaving a server
func (h *MonitorEventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Left server: %v", e))
}

// OnServerPublish ???
func (h *MonitorEventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Server publish to channel %s with data %v", e.Channel, string(e.Data)))
	// todo make this configurable
	//h.onServerPublishLinear(nil, e)
	h.onServerPublishParallel(nil, e)
}

func (h *MonitorEventHandler) processMempoolPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	tx, err := h.monitor.Processor().FilterMempoolPublishEvent(e)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] failed to process server event: %v", err))
		return
	}

	if h.monitor.SaveDestinations() {
		// Process transaction and save outputs
	}

	if tx == "" {
		return
	}
	_, err = h.buxClient.RecordMonitoredTransaction(h.ctx, tx)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v", err))
		return
	}

	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %v", tx))
	}
}

func (h *MonitorEventHandler) processBlockHeaderPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	bi := whatsonchain.BlockInfo{}
	err := json.Unmarshal(e.Data, &bi)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR unmarshalling block header: %v", err))
		return
	}
	bh := bc.BlockHeader{
		HashPrevBlock:  []byte(bi.PreviousBlockHash),
		HashMerkleRoot: []byte(bi.MerkleRoot),
		Nonce:          uint32(bi.Nonce),
		Version:        uint32(bi.Version),
		Time:           uint32(bi.Time),
		Bits:           []byte(bi.Bits),
	}
	_, err = h.buxClient.RecordBlockHeader(h.ctx, bi.Hash, bh)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v", err))
		return
	}

	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] successfully recorded blockheader: %v", bi.Hash))
	}
}

func (h *MonitorEventHandler) onServerPublishLinear(c *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	switch e.Channel {
	case "mempool:transactions":
		h.processMempoolPublish(c, e)
	case "block:headers":
		h.processBlockHeaderPublish(c, e)
	}
	tx, err := h.monitor.Processor().FilterMempoolPublishEvent(e)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] failed to process server event: %v", err))
		return
	}

	if h.monitor.SaveDestinations() {
		// Process transaction and save outputs
	}

	if tx == "" {
		return
	}
	_, err = h.buxClient.RecordMonitoredTransaction(h.ctx, tx)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v", err))
		return
	}

	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %v", tx))
	}
}

func (h *MonitorEventHandler) onServerPublishParallel(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	_, err := h.limit.Execute(func() {
		h.onServerPublishLinear(nil, e)
	})

	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR failed to start goroutine: %v", err))
	}
}

// SetMonitor sets the monitor for the given handler
func (h *MonitorEventHandler) SetMonitor(monitor *chainstate.Monitor) {
	h.monitor = monitor
}

// RecordTransaction records a transaction into bux
func (h *MonitorEventHandler) RecordTransaction(ctx context.Context, txHex string) error {

	_, err := h.buxClient.RecordMonitoredTransaction(ctx, txHex)

	return err
}

// RecordBlockHeader records a block header into bux
func (h *MonitorEventHandler) RecordBlockHeader(ctx context.Context, bh bc.BlockHeader) error {
	return nil
}

// GetWhatsOnChain returns the whats on chain client interface
func (h *MonitorEventHandler) GetWhatsOnChain() whatsonchain.ClientInterface {

	return h.buxClient.Chainstate().WhatsOnChain()
}
