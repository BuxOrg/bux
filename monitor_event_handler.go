package bux

import (
	"context"
	"fmt"
	"runtime"

	"github.com/mrz1836/go-whatsonchain"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
	"github.com/korovkin/limiter"
)

type eventHandler struct {
	monitor   chainstate.MonitorService
	buxClient ClientInterface
	xpub      string
	ctx       context.Context
	limit     *limiter.ConcurrencyLimiter
}

// NewMonitorHandler create a new monitor handler
func NewMonitorHandler(ctx context.Context, xpubKey string, buxClient ClientInterface, monitor chainstate.MonitorService) eventHandler {
	return eventHandler{
		monitor:   monitor,
		buxClient: buxClient,
		xpub:      xpubKey,
		ctx:       ctx,
		limit:     limiter.NewConcurrencyLimiter(runtime.NumCPU()),
	}
}

func (h *eventHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	fmt.Printf("Conntected to server: %s\n", e.ClientID)
	if h.monitor.GetProcessMempoolOnConnect() {
		fmt.Printf("PROCESS MEMPOOL\n")
		go func() {
			err := h.monitor.ProcessMempool(context.Background())
			if err != nil {
				fmt.Printf("ERROR processing mempool: %s\n", err.Error())
			}
		}()
	}
	h.monitor.Connected()
}

func (h *eventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	fmt.Printf("Error: %s", e.Message)
}

func (h *eventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
}

func (h *eventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	h.monitor.Disconnected()
}

func (h *eventHandler) OnJoin(_ *centrifuge.Subscription, e centrifuge.JoinEvent) {
}

func (h *eventHandler) OnLeave(_ *centrifuge.Subscription, e centrifuge.LeaveEvent) {
}

func (h *eventHandler) OnPublish(_ *centrifuge.Subscription, e centrifuge.PublishEvent) {
}

func (h *eventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
}

func (h *eventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
}

func (h *eventHandler) OnSubscribeSuccess(_ *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
}

func (h *eventHandler) OnSubscribeError(_ *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
}

func (h *eventHandler) OnUnsubscribe(_ *centrifuge.Subscription, e centrifuge.UnsubscribeEvent) {
}

func (h *eventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	fmt.Printf("Joined server: %v\n", e)
}

func (h *eventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
}

func (h *eventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	// todo make this configurable
	//h.OnServerPublishLinear(nil, e)
	h.OnServerPublishParallel(nil, e)
}

func (h *eventHandler) OnServerPublishLinear(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	tx, err := h.monitor.Processor().FilterMempoolPublishEvent(e)
	if err != nil {
		fmt.Printf("failed to process server event: %v\n", err)
		return
	}

	if tx == "" {
		return
	}
	_, err = h.buxClient.RecordTransaction(h.ctx, h.xpub, tx, "")
	if err != nil {
		fmt.Printf("error recording tx: %v\n", err)
		return
	}
	fmt.Printf("successfully recorded tx: %v\n", tx)
}

func (h *eventHandler) OnServerPublishParallel(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	_, err := h.limit.Execute(func() {
		h.OnServerPublishLinear(nil, e)
	})

	if err != nil {
		fmt.Printf("failed to start goroutine: %v", err)
	}
}

// SetMonitor sets the monitor for the given handler
func (h *eventHandler) SetMonitor(monitor *chainstate.Monitor) {
	h.monitor = monitor
}

// RecordTransaction records a transaction into bux
func (h *eventHandler) RecordTransaction(ctx context.Context, xPubKey, txHex, draftID string) error {

	_, err := h.buxClient.RecordTransaction(ctx, xPubKey, txHex, draftID)

	return err
}

// GetWhatsOnChain returns the whats on chain client interface
func (h *eventHandler) GetWhatsOnChain() whatsonchain.ClientInterface {

	return h.buxClient.Chainstate().WhatsOnChain()
}
