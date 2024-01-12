package chainstate

import (
	"context"
	"net/http"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoin-sv/go-broadcast-client/broadcast"
	"github.com/centrifugal/centrifuge-go"
	"github.com/tonicpow/go-minercraft/v2"
)

// HTTPInterface is the HTTP client interface
type HTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// ChainService is the chain related methods
type ChainService interface {
	Broadcast(ctx context.Context, id, txHex string, timeout time.Duration) (string, error)
	QueryTransaction(
		ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
	) (*TransactionInfo, error)
	QueryTransactionFastest(
		ctx context.Context, id string, requiredIn RequiredIn, timeout time.Duration,
	) (*TransactionInfo, error)
}

// ProviderServices is the chainstate providers interface
type ProviderServices interface {
	Minercraft() minercraft.ClientInterface
	BroadcastClient() broadcast.Client
}

// HeaderService is header services interface
type HeaderService interface {
	VerifyMerkleRoots(ctx context.Context, merkleRoots []MerkleRootConfirmationRequestItem) error
}

// ClientInterface is the chainstate client interface
type ClientInterface interface {
	ChainService
	ProviderServices
	HeaderService
	Close(ctx context.Context)
	Debug(on bool)
	DebugLog(text string)
	HTTPClient() HTTPInterface
	IsDebug() bool
	IsNewRelicEnabled() bool
	Network() Network
	QueryTimeout() time.Duration
	FeeUnit() *utils.FeeUnit
}

// SocketHandler is composite interface of centrifuge handlers interfaces
type SocketHandler interface {
	OnConnect(*centrifuge.Client, centrifuge.ConnectEvent)
	OnDisconnect(*centrifuge.Client, centrifuge.DisconnectEvent)
	OnError(*centrifuge.Client, centrifuge.ErrorEvent)
	OnJoin(*centrifuge.Subscription, centrifuge.JoinEvent)
	OnLeave(*centrifuge.Subscription, centrifuge.LeaveEvent)
	OnMessage(*centrifuge.Client, centrifuge.MessageEvent)
	OnPublish(*centrifuge.Subscription, centrifuge.PublishEvent)
	OnServerJoin(*centrifuge.Client, centrifuge.ServerJoinEvent)
	OnServerLeave(*centrifuge.Client, centrifuge.ServerLeaveEvent)
	OnServerPublish(*centrifuge.Client, centrifuge.ServerPublishEvent)
	OnServerSubscribe(*centrifuge.Client, centrifuge.ServerSubscribeEvent)
	OnServerUnsubscribe(*centrifuge.Client, centrifuge.ServerUnsubscribeEvent)
	OnSubscribeError(*centrifuge.Subscription, centrifuge.SubscribeErrorEvent)
	OnSubscribeSuccess(*centrifuge.Subscription, centrifuge.SubscribeSuccessEvent)
	OnUnsubscribe(*centrifuge.Subscription, centrifuge.UnsubscribeEvent)
}
