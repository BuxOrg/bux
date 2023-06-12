package bux

import (
	"context"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	centrifuge "github.com/centrifugal/centrifuge-go/0.9.6"
	zLogger "github.com/mrz1836/go-logger"
)

// PulseEventHandler for handling transaction events from a monitor
type PulseEventHandler struct {
	buxClient ClientInterface
	ctx       context.Context
	debug     bool
	logger    chainstate.Logger
	monitor   chainstate.PulseService
}

// NewPulseHandler create a new monitor handler
func NewPulseHandler(ctx context.Context, buxClient ClientInterface, monitor chainstate.PulseService) PulseEventHandler {
	return PulseEventHandler{
		buxClient: buxClient,
		ctx:       ctx,
		logger:    monitor.Logger(),
		monitor:   monitor,
		debug:     monitor.IsDebug(),
	}
}

// OnConnecting event when connecting
func (h *PulseEventHandler) OnConnecting(e centrifuge.ConnectingEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Connecting - %d (%s)", e.Code, e.Reason))
}

// OnConnected event when connected
func (h *PulseEventHandler) OnConnected(e centrifuge.ConnectedEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Connected with ID %s", e.ClientID))
	h.monitor.Connected()
}

// OnDisconnected event when disconnected
func (h *PulseEventHandler) OnDisconnected(e centrifuge.DisconnectedEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Disconnected - %d (%s)", e.Code, e.Reason))
	h.monitor.Disconnected()
}

// OnMessage event when new message received
func (h *PulseEventHandler) OnMessage(e centrifuge.MessageEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Message - %s", e.Data))
	}
}

// OnError event on error
func (h *PulseEventHandler) OnError(e centrifuge.ErrorEvent) {
	h.logger.Error(h.ctx, fmt.Sprintf("[PULSE] Error: %s", e.Error.Error()))
}

// OnServerPublication event on received channel Publication
func (h *PulseEventHandler) OnServerPublication(e centrifuge.ServerPublicationEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Event received: %s (offset %d)", e.Data, e.Offset))
	}
}

// OnServerSubscribed event on server subscribed
func (h *PulseEventHandler) OnServerSubscribed(e centrifuge.ServerSubscribedEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Subscribed: %s", e.Data))
	}
}

// OnServerSubscribing event on server subscribing
func (h *PulseEventHandler) OnServerSubscribing(e centrifuge.ServerSubscribingEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Subscribing - %s", e.Channel))
	}
}

// OnServerUnsubscribed event on server unsubscribed
func (h *PulseEventHandler) OnServerUnsubscribed(e centrifuge.ServerUnsubscribedEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] Unsubscribed - %s", e.Channel))
	}
}

// OnServerJoin event when joining a server
func (h *PulseEventHandler) OnServerJoin(e centrifuge.ServerJoinEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] onJoin - %s", e.ConnInfo))
	}
}

// OnServerLeave event when leaving a server
func (h *PulseEventHandler) OnServerLeave(e centrifuge.ServerLeaveEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE] onLeave - %s", e.ConnInfo))
	}
}

// OnPublication event when leaving a server
func (h *PulseEventHandler) OnPublication(e centrifuge.PublicationEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] Event received: %s (offset %d)", e.Data, e.Offset))

	txs, err := getTransactionsWithoutMerkleProof(context.Background(), nil, "", nil, WithClient(h.buxClient))
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[PULSE Subscription] Error: %s", err.Error()))
		return
	}
	for _, tx := range txs {
		if buff, err := h.buxClient.Chainstate().QueryMAPITransaction(context.Background(), tx.ID, "mempool", 10*time.Second); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[PULSE Subscription] QueryTransaction error: %s", err.Error()))
		} else if buff != nil && buff.MerkleProof != nil {
			tx.MerkleProof = MerkleProof(*buff.MerkleProof)
			err = tx.Save(context.Background())
			if err != nil {
				h.logger.Error(h.ctx, fmt.Sprintf("[PULSE Subscription] Merkle Proof wasn't updated. Error: %s", err.Error()))
			}
		}
	}
}

// OnSubscribing event when subscribing
func (h *PulseEventHandler) OnSubscribing(e centrifuge.SubscribingEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] Subscribing. Status: %d (%s)", e.Code, e.Reason))
}

// OnSubscribed event when subscribed
func (h *PulseEventHandler) OnSubscribed(e centrifuge.SubscribedEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] Subscribed (%v)", e))
}

// OnUnsubscribed event when unsubscribed
func (h *PulseEventHandler) OnUnsubscribed(e centrifuge.UnsubscribedEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] Unsubscribed.Status: %d (%s)", e.Code, e.Reason))
}

// OnSubscriptionError event when subscription error occurs
func (h *PulseEventHandler) OnSubscriptionError(e centrifuge.SubscriptionErrorEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] Subscription error: %s", e.Error))
}

// OnSubscriptionJoin event when someone join to channel
func (h *PulseEventHandler) OnSubscriptionJoin(e centrifuge.JoinEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] onJoin - %s", e.ConnInfo))
	}
}

// OnSubscriptionLeave event when someone left from channel
func (h *PulseEventHandler) OnSubscriptionLeave(e centrifuge.LeaveEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[PULSE Subscription] onLeave - %s", e.ConnInfo))
	}
}

// SetPulse sets the pulse for the given handler
func (h *PulseEventHandler) SetPulse(monitor *chainstate.Pulse) {
	h.monitor = monitor
	h.logger = zLogger.NewGormLogger(false, 4)
	h.debug = monitor.IsDebug()
}
