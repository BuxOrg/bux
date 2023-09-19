package chainstate

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/centrifugal/centrifuge-go"
)

// AddFilterMessage defines a new filter to be published from the client
// todo Just rely on the agent for this data type
type AddFilterMessage struct {
	Filter    string `json:"filter"`
	Hash      string `json:"hash"`
	Regex     string `json:"regex"`
	Timestamp int64  `json:"timestamp"`
}

// SetFilterMessage defines a new filter message with a list of filters
type SetFilterMessage struct {
	Filter    []byte `json:"filter"`
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

// AddFilter adds a new filter to the agent
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

// SetFilter (re)sets a filter to the agent
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

// newCentrifugeClient will create a new Centrifuge using the provided handler and default configurations
func newCentrifugeClient(wsURL string, handler SocketHandler) MonitorClient {
	c := centrifuge.NewJsonClient(wsURL, centrifuge.DefaultConfig()) // todo: use our own defaults/custom options

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
