package bux

import (
	"context"
	"fmt"
	"runtime"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
	"github.com/korovkin/limiter"
	"github.com/mrz1836/go-whatsonchain"
)

type transactionEventHandler struct {
	debug     bool
	logger    chainstate.Logger
	monitor   chainstate.MonitorService
	buxClient ClientInterface
	ctx       context.Context
	limit     *limiter.ConcurrencyLimiter
}

// NewTransactionMonitorHandler create a new monitor handler
func NewTransactionMonitorHandler(ctx context.Context, buxClient ClientInterface, monitor chainstate.MonitorService) transactionEventHandler {
	return transactionEventHandler{
		debug:     monitor.IsDebug(),
		logger:    monitor.Logger(),
		monitor:   monitor,
		buxClient: buxClient,
		ctx:       ctx,
		limit:     limiter.NewConcurrencyLimiter(runtime.NumCPU()),
	}
}

func (h *transactionEventHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	ctx := context.Background()
	h.logger.Info(ctx, fmt.Sprintf("[MONITOR] Connected to server: %s\n", e.ClientID))
	if h.monitor.GetProcessMempoolOnConnect() {
		h.logger.Info(ctx, "[MONITOR] PROCESS MEMPOOL")
		go func() {
			err := h.monitor.ProcessMempool(ctx)
			if err != nil {
				h.logger.Error(ctx, fmt.Sprintf("[MONITOR] ERROR processing mempool: %s", err.Error()))
			}
		}()
	}
	h.monitor.Connected()
}

func (h *transactionEventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	ctx := context.Background()
	h.logger.Error(ctx, fmt.Sprintf("[MONITOR] Error: %s", e.Message))
}

func (h *transactionEventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
}

func (h *transactionEventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	h.monitor.Disconnected()
}

func (h *transactionEventHandler) OnJoin(_ *centrifuge.Subscription, e centrifuge.JoinEvent) {
}

func (h *transactionEventHandler) OnLeave(_ *centrifuge.Subscription, e centrifuge.LeaveEvent) {
}

func (h *transactionEventHandler) OnPublish(_ *centrifuge.Subscription, e centrifuge.PublishEvent) {
}

func (h *transactionEventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
}

func (h *transactionEventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
}

func (h *transactionEventHandler) OnSubscribeSuccess(_ *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
}

func (h *transactionEventHandler) OnSubscribeError(_ *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
}

func (h *transactionEventHandler) OnUnsubscribe(_ *centrifuge.Subscription, e centrifuge.UnsubscribeEvent) {
}

func (h *transactionEventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	ctx := context.Background()
	h.logger.Info(ctx, fmt.Sprintf("[MONITOR] Joined server: %v\n", e))
}

func (h *transactionEventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
}

func (h *transactionEventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	// todo make this configurable
	//h.OnServerPublishLinear(nil, e)
	h.OnServerPublishParallel(nil, e)
}

func (h *transactionEventHandler) OnServerPublishLinear(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	ctx := context.Background()
	tx, err := h.monitor.Processor().FilterMempoolPublishEvent(e)
	if err != nil {
		h.logger.Error(ctx, fmt.Sprintf("[MONITOR] failed to process server event: %v\n", err))
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
		h.logger.Error(ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v\n", err))
		return
	}

	if h.debug {
		h.logger.Info(ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %v\n", tx))
	}
}

func (h *transactionEventHandler) OnServerPublishParallel(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	ctx := context.Background()
	_, err := h.limit.Execute(func() {
		h.OnServerPublishLinear(nil, e)
	})

	if err != nil {
		h.logger.Error(ctx, fmt.Sprintf("[MONITOR] ERROR failed to start goroutine: %v", err))
	}
}

// SetMonitor sets the monitor for the given handler
func (h *transactionEventHandler) SetMonitor(monitor *chainstate.Monitor) {
	h.monitor = monitor
}

// RecordTransaction records a transaction into bux
func (h *transactionEventHandler) RecordTransaction(ctx context.Context, txHex string) error {

	_, err := h.buxClient.RecordMonitoredTransaction(ctx, txHex)

	return err
}

// GetWhatsOnChain returns the whats on chain client interface
func (h *transactionEventHandler) GetWhatsOnChain() whatsonchain.ClientInterface {

	return h.buxClient.Chainstate().WhatsOnChain()
}
