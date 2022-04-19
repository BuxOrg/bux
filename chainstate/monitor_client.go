package chainstate

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/mrz1836/go-whatsonchain"
)

// AddFilterMessage defines a new Filter to be published from the client
// todo Just rely on the agent for this data type
type AddFilterMessage struct {
	Filter    string `json:"Filter"`
	Hash      string `json:"hash"`
	Regex     string `json:"regex"`
	Timestamp int64  `json:"timestamp"`
}

// SetFilterMessage defines a new filter message with a list of filters
type SetFilterMessage struct {
	Filter    []byte `json:"Filter"`
	Hash      string `json:"hash"`
	Regex     string `json:"regex"`
	Timestamp int64  `json:"timestamp"`
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

// SetToken set the client token
func (a *AgentClient) SetToken(token string) {
	a.Client.SetToken(token)
}

// AddFilter adds a new Filter to the agent
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
	return a.Client.Publish("add_filter", data)
}

// SetFilter (re)sets a Filter to the agent
func (a *AgentClient) SetFilter(regex string, bloomFilter *BloomProcessorFilter) (centrifuge.PublishResult, error) {
	filter := new(bytes.Buffer)
	_, err := bloomFilter.Filter.WriteTo(filter)
	if err != nil {
		return centrifuge.PublishResult{}, err
	}

	msg := SetFilterMessage{
		Regex:     regex,
		Filter:    filter.Bytes(),
		Timestamp: time.Now().Unix(),
	}

	var data []byte
	data, err = json.Marshal(msg)
	if err != nil {
		return centrifuge.PublishResult{}, err
	}
	return a.Client.Publish("set_filter", data)
}

func newCentrifugeClient(wsURL string, handler whatsonchain.SocketHandler) MonitorClient {
	c := centrifuge.NewJsonClient(wsURL, centrifuge.DefaultConfig())

	c.OnConnect(handler)
	c.OnDisconnect(handler)
	c.OnError(handler)
	c.OnMessage(handler)
	c.OnServerJoin(handler)
	c.OnServerLeave(handler)
	c.OnServerPublish(handler)
	c.OnServerSubscribe(handler)
	c.OnServerUnsubscribe(handler)

	return &AgentClient{Client: c}
}
