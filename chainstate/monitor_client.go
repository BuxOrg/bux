package chainstate

import (
	"encoding/json"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
)

// MonitorClient interface
type MonitorClient interface {
	Connect() error
	Disconnect() error
	SetToken(token string)
	AddFilter(regex, item string) (centrifuge.PublishResult, error)
}

// AgentClient implements MonitorClient with needed agent methods
type AgentClient struct {
	*centrifuge.Client
	Token string
}

// Connect establishes connection to agent
func (a *AgentClient) Connect() error {
	return a.Client.Connect()
}

// Disconnect closes connection to agent
func (a *AgentClient) Disconnect() error {
	return a.Client.Disconnect()
}

func (a *AgentClient) SetToken(token string) {
	a.Client.SetToken(token)
}

// TODO: Just rely on the agent for this data type
// AddFilterMessage defines a new filter to be published from the client
type AddFilterMessage struct {
	Timestamp int64  `json:"timestamp"`
	Regex     string `json:"regex"`
	Filter    string `json:"filter"`
	Hash      string `json:"hash"`
}

// AddFilter adds a new filtero the agent
func (a *AgentClient) AddFilter(regex, item string) (centrifuge.PublishResult, error) {
	msg := AddFilterMessage{
		Regex:     regex,
		Filter:    item,
		Timestamp: time.Now().Unix(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return centrifuge.PublishResult{}, err
	}
	return a.Client.Publish("filter", data)
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

	return &AgentClient{Client: c}
}
