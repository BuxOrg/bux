package chainstate

import (
	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
)

// MonitorClient interface
type MonitorClient interface {
	Connect() error
	Disconnect() error
}

func newCentrifugeClient(wsURL string, handler whatsonchain.SocketHandler) MonitorClient {
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
