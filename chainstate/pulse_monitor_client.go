package chainstate

import (
	"fmt"

	centrifuge "github.com/centrifugal/centrifuge-go/0.9.6"
)

// PulseAgentClient implements MonitorClient with needed agent methods
type PulseAgentClient struct {
	*centrifuge.Client
	*centrifuge.Subscription
}

// Connect establishes connection to agent
func (a *PulseAgentClient) Connect() error {
	return a.Client.Connect()
}

// Disconnect closes connection to agent
func (a *PulseAgentClient) Disconnect() error {
	return a.Client.Disconnect()
}

// Subscribe subscribes to headers
func (a *PulseAgentClient) Subscribe(handler PulseSubscriptionHandler) error {
	if a.Subscription != nil {
		return fmt.Errorf("Pulse is already subscribed")
	}
	sub, err := a.Client.NewSubscription("headers", centrifuge.SubscriptionConfig{
		Recoverable: true,
		Positioned:  true,
	})
	if err != nil {
		return err
	}
	sub.OnPublication(handler.OnPublication)
	sub.OnSubscribing(handler.OnSubscribing)
	sub.OnSubscribed(handler.OnSubscribed)
	sub.OnUnsubscribed(handler.OnUnsubscribed)
	sub.OnError(handler.OnSubscriptionError)
	sub.OnJoin(handler.OnSubscriptionJoin)
	sub.OnLeave(handler.OnSubscriptionLeave)

	err = sub.Subscribe()
	if err != nil {
		return err
	}
	a.Subscription = sub
	return nil
}

// Unsubscribe unsubscribes
func (a *PulseAgentClient) Unsubscribe() error {
	if a.Subscription == nil {
		return nil
	}
	err := a.Subscription.Unsubscribe()
	if err != nil {
		return err
	}
	a.Subscription = nil
	return nil
}

// newPulseCentrifugeClient will create a new Centrifuge using the provided handler and default configurations
func newPulseCentrifugeClient(wsURL string, token string, handler PulseHandler) PulseMonitorClient {
	cfg := centrifuge.Config{}
	if token != "" {
		cfg.Token = token
	}
	c := centrifuge.NewJsonClient(wsURL, cfg)

	c.OnConnecting(handler.OnConnecting)
	c.OnConnected(handler.OnConnected)
	c.OnError(handler.OnError)
	c.OnMessage(handler.OnMessage)
	c.OnJoin(handler.OnServerJoin)
	c.OnLeave(handler.OnServerLeave)
	c.OnPublication(handler.OnServerPublication)
	c.OnSubscribed(handler.OnServerSubscribed)
	c.OnSubscribing(handler.OnServerSubscribing)

	return &PulseAgentClient{Client: c}
}
