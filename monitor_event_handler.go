package bux

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/centrifugal/centrifuge-go"
	"github.com/korovkin/limiter"
	"github.com/libsv/go-bc"
	"github.com/mrz1836/go-whatsonchain"
)

// MonitorEventHandler for handling transaction events from a monitor
type MonitorEventHandler struct {
	blockSyncChannel    chan bool
	buxClient           ClientInterface
	ctx                 context.Context
	debug               bool
	limit               *limiter.ConcurrencyLimiter
	logger              chainstate.Logger
	monitor             chainstate.MonitorService
}

type blockSubscriptionHandler struct {
	buxClient ClientInterface
	ctx       context.Context
	debug     bool
	logger    chainstate.Logger
	monitor   chainstate.MonitorService
	wg        sync.WaitGroup
}

func (b *blockSubscriptionHandler) OnPublish(subscription *centrifuge.Subscription, e centrifuge.PublishEvent) {

	channelName := subscription.Channel()
	if strings.HasPrefix(channelName, "block:sync:") {
		// block subscription
		tx, err := b.monitor.Processor().FilterTransactionPublishEvent(e.Data)
		if err != nil {
			b.logger.Error(b.ctx, fmt.Sprintf("[MONITOR] Error processing block data: %s", err.Error()))
		}

		if tx == "" {
			return
		}

		if _, err = recordMonitoredTransaction(b.ctx, b.buxClient, tx); err != nil {
			b.logger.Error(b.ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v", err))
			return
		}

		if b.debug {
			b.logger.Info(b.ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %v", tx))
		}
	}
}

func (b *blockSubscriptionHandler) OnUnsubscribe(subscription *centrifuge.Subscription, _ centrifuge.UnsubscribeEvent) {

	b.logger.Info(b.ctx, fmt.Sprintf("[MONITOR] OnUnsubscribe: %s", subscription.Channel()))
	// close wait group
	b.wg.Done()
}

// NewMonitorHandler create a new monitor handler
func NewMonitorHandler(ctx context.Context, buxClient ClientInterface, monitor chainstate.MonitorService) MonitorEventHandler {
	return MonitorEventHandler{
		buxClient: buxClient,
		ctx:       ctx,
		debug:     monitor.IsDebug(),
		limit:     limiter.NewConcurrencyLimiter(runtime.NumCPU()),
		logger:    monitor.Logger(),
		monitor:   monitor,
	}
}

// OnConnect event when connected
func (h *MonitorEventHandler) OnConnect(client *centrifuge.Client, e centrifuge.ConnectEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Connected to server: %s", e.ClientID))

	agentClient := &chainstate.AgentClient{
		Client: client,
	}
	filters := h.monitor.Processor().GetFilters()
	for regex, bloomFilter := range filters {
		if _, err := agentClient.SetFilter(regex, bloomFilter); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR processing mempool: %s", err.Error()))
		}
	}

		if h.monitor.GetProcessMempoolOnConnect() {
			h.logger.Info(h.ctx, "[MONITOR] PROCESS MEMPOOL")
			go func() {
				if err := h.monitor.ProcessMempool(h.ctx); err != nil {
					h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR processing mempool: %s", err.Error()))
				}
			}()
		}

		h.logger.Info(h.ctx, "[MONITOR] PROCESS BLOCK HEADERS")
		if err := h.ProcessBlockHeaders(h.ctx, client); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR processing block headers: %s", err.Error()))
		}

	h.logger.Info(h.ctx, "[MONITOR] PROCESS BLOCKS")
	h.blockSyncChannel = make(chan bool)
	go func() {
		if err := h.ProcessBlocks(h.ctx, client, h.blockSyncChannel); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR processing blocks: %s", err.Error()))
		}
	}()

	h.monitor.Connected()
}

// ProcessBlocks processes all transactions in blocks that have not yet been synced
func (h *MonitorEventHandler) ProcessBlocks(ctx context.Context, client *centrifuge.Client, blockChannel chan bool) error {
	h.logger.Info(ctx, "[MONITOR] ProcessBlocks start")
	for {
		// Check if channel has been closed
		select {
		case <-blockChannel:
			h.logger.Info(ctx, "[MONITOR] block sync channel closed, stopping ProcessBlocks")
			return nil
		default:
			// get all block headers that have not been marked as synced
			blockHeaders, err := h.buxClient.GetUnsyncedBlockHeaders(ctx)
			if err != nil {
				h.logger.Error(ctx, err.Error())
			} else {
				h.logger.Info(ctx, fmt.Sprintf("[MONITOR] processing block headers: %d", len(blockHeaders)))
				for _, blockHeader := range blockHeaders {
					h.logger.Info(ctx, fmt.Sprintf("[MONITOR] Processing block %d: %s", blockHeader.Height, blockHeader.ID))
					handler := &blockSubscriptionHandler{
						buxClient: h.buxClient,
						ctx:       ctx,
						debug:     h.debug,
						logger:    h.logger,
						monitor:   h.monitor,
					}
					handler.wg.Add(1)

					var subscription *centrifuge.Subscription
					subscription, err = client.NewSubscription("block:sync:" + blockHeader.ID)
					if err != nil {
						h.logger.Error(ctx, err.Error())
					} else {
						h.logger.Info(ctx, fmt.Sprintf("[MONITOR] Starting block subscription: %v", subscription))
						subscription.OnPublish(handler)
						subscription.OnUnsubscribe(handler)

						if err = subscription.Subscribe(); err != nil {
							h.logger.Error(ctx, err.Error())
							handler.wg.Done()
						} else {
							h.logger.Info(ctx, "[MONITOR] Waiting for wait group to finish")
							handler.wg.Wait()

							// save that block header has been synced
							blockHeader.Synced.Valid = true
							blockHeader.Synced.Time = time.Now()
							if err = blockHeader.Save(ctx); err != nil {
								h.logger.Error(ctx, err.Error())
							}
						}
					}
				}
			}

			time.Sleep(defaultSleepForNewBlockHeaders)
		}
	}
}

// ProcessBlockHeaders processes all missing block headers
func (h *MonitorEventHandler) ProcessBlockHeaders(ctx context.Context, client *centrifuge.Client) error {

	lastBlockHeader, err := h.buxClient.GetLastBlockHeader(ctx)
	if err != nil {
		h.logger.Error(h.ctx, err.Error())
		return err
	}
	if lastBlockHeader == nil {
		h.logger.Info(h.ctx, "no last block header found, skipping...")
		return nil
	}
	var subscription *centrifuge.Subscription
	subscription, err = client.NewSubscription("block:headers:history:" + fmt.Sprint(lastBlockHeader.Height))
	if err != nil {
		h.logger.Error(h.ctx, err.Error())
	} else {
		h.logger.Info(ctx, fmt.Sprintf("[MONITOR] Starting block header subscription: %v", subscription))
		subscription.OnPublish(h)
		if err = subscription.Subscribe(); err != nil {
			h.logger.Error(h.ctx, err.Error())
		}
	}

	return nil
}

// OnError on error event
func (h *MonitorEventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] Error: %s", e.Message))
}

// OnMessage on new message event
func (h *MonitorEventHandler) OnMessage(_ *centrifuge.Client, e centrifuge.MessageEvent) {
	var data map[string]interface{}
	err := json.Unmarshal(e.Data, &data)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] failed unmarshalling data: %s", err.Error()))
	}

	if _, ok := data["time"]; !ok {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnMessage: %v", data))
	}
}

// OnDisconnect when disconnected
func (h *MonitorEventHandler) OnDisconnect(_ *centrifuge.Client, _ centrifuge.DisconnectEvent) {
	defer close(h.blockSyncChannel)

	defer func(logger chainstate.Logger) {
		ctx := context.Background()
		rec := recover()
		if rec != nil {
			logger.Error(ctx, fmt.Sprintf("[MONITOR] Tried closing a closed channel: %v", rec))
		}
	}(h.logger)

	h.monitor.Disconnected()
}

// OnJoin event when joining a server
func (h *MonitorEventHandler) OnJoin(_ *centrifuge.Subscription, e centrifuge.JoinEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnJoin: %v", e))
	}
}

// OnLeave event when leaving a server
func (h *MonitorEventHandler) OnLeave(_ *centrifuge.Subscription, e centrifuge.LeaveEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnLeave: %v", e))
	}
}

// OnPublish on publish event
func (h *MonitorEventHandler) OnPublish(subscription *centrifuge.Subscription, e centrifuge.PublishEvent) {
	channelName := subscription.Channel()
	if strings.HasPrefix(channelName, "block:sync:") {
		// block subscription
		tx, err := h.monitor.Processor().FilterTransactionPublishEvent(e.Data)
		if err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR filtering tx: %v", err))
			return
		}

		if tx == "" {
			return
		}
		if _, err = recordMonitoredTransaction(h.ctx, h.buxClient, tx); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v", err))
			return
		}

		if h.debug {
			h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %v", tx))
		}
	} else if strings.HasPrefix(channelName, "block:headers:history:") {
		bi := whatsonchain.BlockInfo{}
		err := json.Unmarshal(e.Data, &bi)
		if err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR unmarshalling block header: %v", err))
		}

		var existingBlock *BlockHeader
		if existingBlock, err = h.buxClient.GetBlockHeaderByHeight(
			h.ctx, uint32(bi.Height),
		); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR getting block header by height: %v", err))
		}
		if existingBlock == nil {
			if err != nil {
				h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR unmarshalling block header: %v", err))
				return
			}
			bh := bc.BlockHeader{
				Bits:           []byte(bi.Bits),
				HashMerkleRoot: []byte(bi.MerkleRoot),
				HashPrevBlock:  []byte(bi.PreviousBlockHash),
				Nonce:          uint32(bi.Nonce),
				Time:           uint32(bi.Time),
				Version:        uint32(bi.Version),
			}
			if _, err = h.buxClient.RecordBlockHeader(
				h.ctx, bi.Hash, uint32(bi.Height), bh,
			); err != nil {
				h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording block header: %v", err))
				return
			}
		}
	} else {
		if h.debug {
			h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnPublish: %v", e.Data))
		}
	}
}

// OnServerSubscribe on server subscribe event
func (h *MonitorEventHandler) OnServerSubscribe(_ *centrifuge.Client, e centrifuge.ServerSubscribeEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnServerSubscribe: %v", e))
	}
}

// OnServerUnsubscribe on the unsubscribe event
func (h *MonitorEventHandler) OnServerUnsubscribe(_ *centrifuge.Client, e centrifuge.ServerUnsubscribeEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnServerUnsubscribe: %v", e))
	}
}

// OnSubscribeSuccess on subscribe success
func (h *MonitorEventHandler) OnSubscribeSuccess(_ *centrifuge.Subscription, e centrifuge.SubscribeSuccessEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnSubscribeSuccess: %v", e))
	}
}

// OnSubscribeError is for an error
func (h *MonitorEventHandler) OnSubscribeError(_ *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] OnSubscribeError: %v", e))
}

// OnUnsubscribe will unsubscribe
func (h *MonitorEventHandler) OnUnsubscribe(_ *centrifuge.Subscription, e centrifuge.UnsubscribeEvent) {
	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] OnUnsubscribe: %v", e))
	}
}

// OnServerJoin event when joining a server
func (h *MonitorEventHandler) OnServerJoin(_ *centrifuge.Client, e centrifuge.ServerJoinEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Joined server: %v", e))
}

// OnServerLeave event when leaving a server
func (h *MonitorEventHandler) OnServerLeave(_ *centrifuge.Client, e centrifuge.ServerLeaveEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Left server: %v", e))
}

// OnServerPublish on server publish event
func (h *MonitorEventHandler) OnServerPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] Server publish to channel %s with data %v", e.Channel, string(e.Data)))
	// todo make this configurable
	// h.onServerPublishLinear(nil, e)
	h.onServerPublishParallel(nil, e)
}

func (h *MonitorEventHandler) processMempoolPublish(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	tx, err := h.monitor.Processor().FilterTransactionPublishEvent(e.Data)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] failed to process server event: %v", err))
		return
	}

	if h.monitor.SaveDestinations() {
		// Process transaction and save outputs
		// todo: replace printf
		fmt.Printf("Should save the destination here...\n")
	}

	if tx == "" {
		return
	}
	if _, err = recordMonitoredTransaction(h.ctx, h.buxClient, tx); err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording tx: %v", err))
		return
	}

	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] successfully recorded tx: %v", tx))
	}
}

func (h *MonitorEventHandler) processBlockHeaderPublish(client *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	bi := whatsonchain.BlockInfo{}
	err := json.Unmarshal(e.Data, &bi)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR unmarshalling block header: %v", err))
		return
	}
	bh := bc.BlockHeader{
		HashPrevBlock:  []byte(bi.PreviousBlockHash),
		HashMerkleRoot: []byte(bi.MerkleRoot),
		Nonce:          uint32(bi.Nonce),
		Version:        uint32(bi.Version),
		Time:           uint32(bi.Time),
		Bits:           []byte(bi.Bits),
	}

	height := uint32(bi.Height)
	var previousBlockHeader *BlockHeader
	previousBlockHeader, err = getBlockHeaderByHeight(h.ctx, height-1, h.buxClient.DefaultModelOptions()...)
	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR retreiving previous block header: %v", err))
		return
	}
	if previousBlockHeader == nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR Previous block header not found: %d", height-1))
		if err = h.ProcessBlockHeaders(h.ctx, client); err != nil {
			h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR processing block headers: %s", err.Error()))
		}
		return
	}

	if _, err = h.buxClient.RecordBlockHeader(h.ctx, bi.Hash, height, bh); err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR recording block header: %v", err))
		return
	}

	if h.debug {
		h.logger.Info(h.ctx, fmt.Sprintf("[MONITOR] successfully recorded blockheader: %v", bi.Hash))
	}
}

func (h *MonitorEventHandler) onServerPublishLinear(c *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	switch e.Channel {
	case "mempool:transactions":
		h.processMempoolPublish(c, e)
	case "block:headers":
		h.processBlockHeaderPublish(c, e)
	}
}

func (h *MonitorEventHandler) onServerPublishParallel(_ *centrifuge.Client, e centrifuge.ServerPublishEvent) {
	_, err := h.limit.Execute(func() {
		h.onServerPublishLinear(nil, e)
	})

	if err != nil {
		h.logger.Error(h.ctx, fmt.Sprintf("[MONITOR] ERROR failed to start goroutine: %v", err))
	}
}

// SetMonitor sets the monitor for the given handler
func (h *MonitorEventHandler) SetMonitor(monitor *chainstate.Monitor) {
	h.monitor = monitor
}

// RecordTransaction records a transaction into bux
func (h *MonitorEventHandler) RecordTransaction(ctx context.Context, txHex string) error {
	_, err := recordMonitoredTransaction(ctx, h.buxClient, txHex)
	return err
}

// RecordBlockHeader records a block header into bux
func (h *MonitorEventHandler) RecordBlockHeader(_ context.Context, _ bc.BlockHeader) error {
	return nil
}

// GetWhatsOnChain returns the WhatsOnChain client interface
func (h *MonitorEventHandler) GetWhatsOnChain() whatsonchain.ClientInterface {
	return h.buxClient.Chainstate().WhatsOnChain()
}
