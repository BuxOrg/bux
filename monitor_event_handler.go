package bux

import (
	"context"
	"fmt"
	"runtime"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
	"github.com/korovkin/limiter"
)

type eventHandler struct {
	monitor   chainstate.MonitorService
	buxClient *Client
	xpub      string
	ctx       context.Context
	limit     *limiter.ConcurrencyLimiter
}

// NewMonitorHandler create a new monitor handler
func NewMonitorHandler(ctx context.Context, xpubKey string, buxClient *Client, monitor chainstate.MonitorService) eventHandler {
	return eventHandler{
		monitor:   monitor,
		buxClient: buxClient,
		xpub:      xpubKey,
		ctx:       ctx,
		limit:     limiter.NewConcurrencyLimiter(runtime.NumCPU()),
	}
}

func (h *eventHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	h.monitor.Connected()
	return
}

func (h *eventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	fmt.Printf("Error: %s", e.Message)
}

func (h *eventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	return
}

func (h *eventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	h.monitor.Disconnected()
	return
}

func (h *eventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
	return
}

func (h *eventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
	return
}

func (h *eventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	return
}

func (h *eventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	return
}

func (h *eventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	// todo make this configurable
	//h.OnServerPublishLinear(nil, e)
	h.OnServerPublishParallel(nil, e)
}

func (h *eventHandler) OnServerPublishLinear(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	tx, err := h.monitor.Processor().FilterMempoolPublishEvent(e)
	if err != nil {
		fmt.Printf("failed to process server event: %v", err)
		return
	}

	if tx == "" {
		return
	}
	_, err = h.buxClient.RecordTransaction(h.ctx, h.xpub, tx, "")
	if err != nil {
		fmt.Printf("error recording tx: %v", err)
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
