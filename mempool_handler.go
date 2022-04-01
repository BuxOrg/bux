package bux

import (
	"context"
	"fmt"
	"log"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
)

type MempoolHandler struct {
	Processor *chainstate.Processor
	BuxClient *Client
}

func PubKeyHashMempoolHandler(client *Client) *MempoolHandler {
	processor := chainstate.PubKeyHashProcessor()
	return &MempoolHandler{
		Processor: processor,
		BuxClient: client,
	}
}

func (h *MempoolHandler) OnConnect(_ *centrifuge.Client, _ centrifuge.ConnectEvent) {
	log.Println("Connected")
}

func (h *MempoolHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	log.Println("Error", e.Message)
}

func (h *MempoolHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	log.Println("Disconnected", e.Reason)
}

func (h *MempoolHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	log.Println(fmt.Sprintf("New message received from channel: %s", string(e.Data)))
}
func (h *MempoolHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	ctx := context.Background()
	tx, err := h.Processor.FilterMempoolPublishEvent(e)
	if err != nil {
		log.Printf("error processing event: %v", err)
	}
	if tx == nil {
		return
	}
	log.Printf("filter accepted transaction: %#v", tx.GetTxID())

	for _, out := range tx.Outputs {
		_, err := h.BuxClient.NewDestinationForLockingScript(ctx, string(h.Processor.FilterType), out.LockingScript.ToString(), true, nil)
		if err != nil {
			log.Printf("error: failed to save destination: %v", err)
			return
		}

	}

	_, err = h.BuxClient.RecordTransaction(ctx, string(h.Processor.FilterType), tx.ToString(), "")
	if err != nil {
		log.Printf("error: failed to record transaction: %v", err)
		return
	}
}
