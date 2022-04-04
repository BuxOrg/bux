package bux

import (
	"context"
	"fmt"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
)

type eventHandler struct {
	monitor   chainstate.MonitorService
	buxClient *Client
	xpub      string
	ctx       context.Context
}

func NewMonitorHandler(ctx context.Context, xpubKey string, buxClient *Client, monitor chainstate.MonitorService) eventHandler {
	return eventHandler{
		monitor:   monitor,
		buxClient: buxClient,
		xpub:      xpubKey,
		ctx:       ctx,
	}
}

func (h *eventHandler) OnConnect(_ *centrifuge.Client, e centrifuge.ConnectEvent) {
	fmt.Printf("Connected to chat with ID %s", e.ClientID)
}

func (h *eventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	fmt.Printf("Error: %s", e.Message)
}

func (h *eventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	fmt.Printf("Message from server: %s", string(e.Data))
	// register transaction
	// h.monitorConfig.
}

func (h *eventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	fmt.Printf("Disconnected from chat: %s", e.Reason)
}

func (h *eventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
	fmt.Printf("Subscribe to server-side channel %s: (resubscribe: %t, recovered: %t)", e.Channel, e.Resubscribed, e.Recovered)
}

func (h *eventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
	fmt.Printf("Unsubscribe from server-side channel %s", e.Channel)
}

func (h *eventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	fmt.Printf("Server-side join to channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *eventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	fmt.Printf("Server-side leave from channel %s: %s (%s)", e.Channel, e.User, e.Client)
}

func (h *eventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	tx, err := h.monitor.Processor().FilterMempoolPublishEvent(e)
	if err != nil {
		fmt.Printf("failed to process server event: %v", err)
		return
	}

	if tx == "" {
		fmt.Printf("filtered transaction...\n")
		return
	}
	_, err = h.buxClient.RecordTransaction(h.ctx, h.xpub, tx, "")
	if err != nil {
		fmt.Printf("error recording tx: %v", err)
		return
	}
	fmt.Printf("successfully recorded tx: %v\n", tx)
	return
}

func (h *eventHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	/*var chatMessage *ChatMessage
	err := json.Unmarshal(e.Data, &chatMessage)
	if err != nil {
		return
	}*/
	fmt.Printf("Someone says via channel %s: %s", sub.Channel(), e.Data)
}

func (h *eventHandler) OnJoin(sub *centrifuge.Subscription, e centrifuge.JoinEvent) {
	fmt.Printf("Someone joined %s: user id %s, client id %s", sub.Channel(), e.User, e.Client)
}

func (h *eventHandler) OnLeave(sub *centrifuge.Subscription, e centrifuge.LeaveEvent) {
	fmt.Printf("Someone left %s: user id %s, client id %s", sub.Channel(), e.User, e.Client)
}

func (h *eventHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	fmt.Printf("Subscribed on channel %s, resubscribed: %v, recovered: %v", sub.Channel(), e.Resubscribed, e.Recovered)
}

func (h *eventHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	fmt.Printf("Subscribed on channel %s failed, error: %s", sub.Channel(), e.Error)
}

func (h *eventHandler) OnUnsubscribe(sub *centrifuge.Subscription, _ centrifuge.UnsubscribeEvent) {
	fmt.Printf("Unsubscribed from channel %s", sub.Channel())
}

func (h *eventHandler) SetMonitor(monitor *chainstate.Monitor) {
	h.monitor = monitor
}
