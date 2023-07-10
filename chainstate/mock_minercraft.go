package chainstate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-bt/v2"
	"github.com/tonicpow/go-minercraft"
)

var (
	minerTaal = &minercraft.Miner{
		MinerID: "030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e",
		Name:    "Taal",
		URL:     "https://merchantapi.taal.com",
	}

	minerMempool = &minercraft.Miner{
		MinerID: "03e92d3e5c3f7bd945dfbf48e7a99393b1bfb3f11f380ae30d286e7ff2aec5a270",
		Name:    "Mempool",
		URL:     "https://www.ddpurse.com/openapi",
		Token:   "561b756d12572020ea9a104c3441b71790acbbce95a6ddbf7e0630971af9424b",
	}

	minerMatterPool = &minercraft.Miner{
		MinerID: "0253a9b2d017254b91704ba52aad0df5ca32b4fb5cb6b267ada6aefa2bc5833a93",
		Name:    "Matterpool",
		URL:     "https://merchantapi.matterpool.io",
	}

	minerGorillaPool = &minercraft.Miner{
		MinerID: "03ad780153c47df915b3d2e23af727c68facaca4facd5f155bf5018b979b9aeb83",
		Name:    "GorillaPool",
		URL:     "https://merchantapi.gorillapool.io",
	}

	allMiners = []*minercraft.Miner{
		minerTaal,
		minerMempool,
		minerGorillaPool,
		minerMatterPool,
	}
)

type MinerCraftBase struct{}

func (m *MinerCraftBase) AddMiner(miner minercraft.Miner) error {
	existingMiner := m.MinerByName(miner.Name)
	if existingMiner != nil {
		return fmt.Errorf("miner %s already exists", miner.Name)
	}
	// Append the new miner
	allMiners = append(allMiners, &miner)
	return nil
}

func (m *MinerCraftBase) BestQuote(context.Context, string, string) (*minercraft.FeeQuoteResponse, error) {
	return nil, nil
}

func (m *MinerCraftBase) FastestQuote(context.Context, time.Duration) (*minercraft.FeeQuoteResponse, error) {
	return nil, nil
}

func (m *MinerCraftBase) FeeQuote(context.Context, *minercraft.Miner) (*minercraft.FeeQuoteResponse, error) {
	return &minercraft.FeeQuoteResponse{
		Quote: &minercraft.FeePayload{
			Fees: []*bt.Fee{
				{
					FeeType:   bt.FeeTypeData,
					MiningFee: bt.FeeUnit(*DefaultFee),
				},
			},
		},
	}, nil
}

func (m *MinerCraftBase) MinerByID(minerID string) *minercraft.Miner {
	for index, miner := range allMiners {
		if strings.EqualFold(minerID, miner.MinerID) {
			return allMiners[index]
		}
	}
	return nil
}

func (m *MinerCraftBase) MinerByName(name string) *minercraft.Miner {
	for index, miner := range allMiners {
		if strings.EqualFold(name, miner.Name) {
			return allMiners[index]
		}
	}
	return nil
}

func (m *MinerCraftBase) Miners() []*minercraft.Miner {
	return allMiners
}

func (m *MinerCraftBase) MinerUpdateToken(name, token string) {
	if miner := m.MinerByName(name); miner != nil {
		miner.Token = token
	}
}

func (m *MinerCraftBase) PolicyQuote(context.Context, *minercraft.Miner) (*minercraft.PolicyQuoteResponse, error) {
	return nil, nil
}

func (m *MinerCraftBase) QueryTransaction(context.Context, *minercraft.Miner, string, ...minercraft.QueryTransactionOptFunc) (*minercraft.QueryTransactionResponse, error) {
	return nil, nil
}

func (m *MinerCraftBase) RemoveMiner(miner *minercraft.Miner) bool {
	for i, cm := range allMiners {
		if cm.Name == miner.Name || cm.MinerID == miner.MinerID {
			allMiners[i] = allMiners[len(allMiners)-1]
			allMiners = allMiners[:len(allMiners)-1]
			return true
		}
	}
	// Miner not found
	return false
}

func (m *MinerCraftBase) SubmitTransaction(context.Context, *minercraft.Miner, *minercraft.Transaction) (*minercraft.SubmitTransactionResponse, error) {
	return nil, nil
}

func (m *MinerCraftBase) SubmitTransactions(context.Context, *minercraft.Miner, []minercraft.Transaction) (*minercraft.SubmitTransactionsResponse, error) {
	return nil, nil
}

func (m *MinerCraftBase) UserAgent() string {
	return "default-user-agent"
}

type minerCraftTxOnChain struct {
	MinerCraftBase
}

func (m *minerCraftTxOnChain) SubmitTransaction(_ context.Context, miner *minercraft.Miner,
	_ *minercraft.Transaction,
) (*minercraft.SubmitTransactionResponse, error) {
	if miner.Name == minercraft.MinerTaal {
		sig := "30440220008615778c5b8610c29b12925c8eb479f692ad6de9e62b7e622a3951baf9fbd8022014aaa27698cd3aba4144bfd707f3323e12ac20101d6e44f22eb8ed0856ef341a"
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.4.0\",\"timestamp\":\"2022-02-01T15:19:40.889523Z\",\"txid\":\"683e11d4db8a776e293dc3bfe446edf66cf3b145a6ec13e1f5f1af6bb5855364\",\"returnResult\":\"failure\",\"resultDescription\":\"Missing inputs\",\"minerId\":\"030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e\",\"currentHighestBlockHash\":\"00000000000000000652def5827ad3de6380376f8fc8d3e835503095a761e0d2\",\"currentHighestBlockHeight\":724807,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "1.4.0",
				CurrentHighestBlockHash:   "00000000000000000652def5827ad3de6380376f8fc8d3e835503095a761e0d2",
				CurrentHighestBlockHeight: 724807,
				MinerID:                   miner.MinerID,
				ResultDescription:         "Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T15:19:40.889523Z",
				TxID:                      onChainExample1TxID,
			},
		}, nil
	} else if miner.Name == minercraft.MinerMempool {
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:  "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-01T17:47:52.518Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Missing inputs\",\"minerId\":null,\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
					Encoding: utf8Type,
					MimeType: applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "",
				CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
				CurrentHighestBlockHeight: 724816,
				MinerID:                   miner.MinerID,
				ResultDescription:         "ERROR: Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T17:47:52.518Z",
				TxID:                      "",
			},
		}, nil
	} else if miner.Name == minercraft.MinerMatterpool {
		sig := matterCloudSig1
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.1.0-1-g35ba2d3\",\"timestamp\":\"2022-02-01T17:50:15.130Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Missing inputs\",\"minerId\":\"0253a9b2d017254b91704ba52aad0df5ca32b4fb5cb6b267ada6aefa2bc5833a93\",\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "1.1.0-1-g35ba2d3",
				CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
				CurrentHighestBlockHeight: 724816,
				MinerID:                   miner.MinerID,
				ResultDescription:         "ERROR: Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T17:50:15.130Z",
				TxID:                      "",
			},
		}, nil
	} else if miner.Name == minercraft.MinerGorillaPool {
		sig := gorillaPoolSig1
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-01T17:52:04.405Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Missing inputs\",\"minerId\":\"03ad780153c47df915b3d2e23af727c68facaca4facd5f155bf5018b979b9aeb83\",\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "",
				CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
				CurrentHighestBlockHeight: 724816,
				MinerID:                   miner.MinerID,
				ResultDescription:         "ERROR: Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T17:52:04.405Z",
				TxID:                      "",
			},
		}, nil
	}

	return nil, errors.New("missing miner response")
}

func (m *minerCraftTxOnChain) QueryTransaction(_ context.Context, miner *minercraft.Miner,
	txID string, _ ...minercraft.QueryTransactionOptFunc,
) (*minercraft.QueryTransactionResponse, error) {
	if txID == onChainExample1TxID && miner.Name == minerTaal.Name {
		sig := "304402207ede387e82db1ac38e4286b0a967b4fe1c8446c413b3785ccf86b56009439b39022043931eae02d7337b039f109be41dbd44d0472abd10ed78d7e434824ea8ab01da"
		pubKey := minerTaal.MinerID
		return &minercraft.QueryTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     minerTaal,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.4.0\",\"timestamp\":\"2022-01-23T19:42:18.6860061Z\",\"txid\":\"908c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b\",\"returnResult\":\"success\",\"blockHash\":\"0000000000000000015122781ab51d57b26a09518630b882f67f1b08d841979d\",\"blockHeight\":723229,\"confirmations\":319,\"minerId\":\"030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e\",\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Query: &minercraft.QueryPayload{
				APIVersion:            "1.4.0",
				Timestamp:             "2022-01-23T19:42:18.6860061Z",
				TxID:                  onChainExample1TxID,
				ReturnResult:          mAPISuccess,
				ResultDescription:     "",
				BlockHash:             onChainExample1BlockHash,
				BlockHeight:           onChainExample1BlockHeight,
				MinerID:               minerTaal.MinerID,
				Confirmations:         onChainExample1Confirmations,
				TxSecondMempoolExpiry: 0,
			},
		}, nil
	} else if txID == onChainExample1TxID && miner.Name == minerMempool.Name {
		return &minercraft.QueryTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     minerMempool,
				Validated: false,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"\",\"timestamp\":\"2022-01-23T19:51:10.046Z\",\"txid\":\"908c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b\",\"returnResult\":\"success\",\"resultDescription\":\"\",\"blockHash\":\"0000000000000000015122781ab51d57b26a09518630b882f67f1b08d841979d\",\"blockHeight\":723229,\"confirmations\":321,\"minerId\":null,\"txSecondMempoolExpiry\":0}",
					Signature: nil, // NOTE: missing from mempool response
					PublicKey: nil, // NOTE: missing from mempool response
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Query: &minercraft.QueryPayload{
				APIVersion:            "", // NOTE: missing from mempool response
				Timestamp:             "2022-01-23T19:51:10.046Z",
				TxID:                  onChainExample1TxID,
				ReturnResult:          mAPISuccess,
				ResultDescription:     "",
				BlockHash:             onChainExample1BlockHash,
				BlockHeight:           onChainExample1BlockHeight,
				MinerID:               "", // NOTE: missing from mempool response
				Confirmations:         onChainExample1Confirmations,
				TxSecondMempoolExpiry: 0,
			},
		}, nil
	}

	return nil, nil
}

type minerCraftBroadcastSuccess struct {
	MinerCraftBase
}

func (m *minerCraftBroadcastSuccess) SubmitTransaction(_ context.Context, miner *minercraft.Miner,
	_ *minercraft.Transaction,
) (*minercraft.SubmitTransactionResponse, error) {
	if miner.Name == minercraft.MinerTaal {
		sig := "30440220268ad023bbe03c62a953f907f81c01754f34ffe4822bb9e89c5245613bda7b7602204c201e56b27fd044b3f8ad77ec2c24dc2b9571166a9a998c256d3cbf598fbbda"
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.4.0\",\"timestamp\":\"2022-02-02T12:12:02.6089293Z\",\"txid\":\"15d31d00ed7533a83d7ab206115d7642812ec04a2cbae4248365febb82576ff3\",\"returnResult\":\"success\",\"resultDescription\":\"\",\"minerId\":\"030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e\",\"currentHighestBlockHash\":\"000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0\",\"currentHighestBlockHeight\":724922,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "1.4.0",
				CurrentHighestBlockHash:   "000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0",
				CurrentHighestBlockHeight: 724922,
				MinerID:                   miner.MinerID,
				ResultDescription:         "",
				ReturnResult:              mAPISuccess,
				Timestamp:                 "2022-02-02T12:12:02.6089293Z",
				TxID:                      broadcastExample1TxID,
			},
		}, nil
	} else if miner.Name == minercraft.MinerMempool {
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:  "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-02T12:12:02.6089293Z\",\"txid\":\"15d31d00ed7533a83d7ab206115d7642812ec04a2cbae4248365febb82576ff3\",\"returnResult\":\"success\",\"resultDescription\":\"\",\"minerId\":\"03e92d3e5c3f7bd945dfbf48e7a99393b1bfb3f11f380ae30d286e7ff2aec5a270\",\"currentHighestBlockHash\":\"000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0\",\"currentHighestBlockHeight\":724922,\"txSecondMempoolExpiry\":0}",
					Encoding: utf8Type,
					MimeType: applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "",
				CurrentHighestBlockHash:   "000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0",
				CurrentHighestBlockHeight: 724922,
				MinerID:                   miner.MinerID,
				ResultDescription:         "",
				ReturnResult:              mAPISuccess,
				Timestamp:                 "2022-02-01T17:47:52.518Z",
				TxID:                      broadcastExample1TxID,
			},
		}, nil
	} else if miner.Name == minercraft.MinerMatterpool {
		sig := matterCloudSig1
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.1.0-1-g35ba2d3\",\"timestamp\":\"2022-02-02T12:12:02.6089293Z\",\"txid\":\"\",\"returnResult\":\"success\",\"resultDescription\":\"\",\"minerId\":\"0253a9b2d017254b91704ba52aad0df5ca32b4fb5cb6b267ada6aefa2bc5833a93\",\"currentHighestBlockHash\":\"000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0\",\"currentHighestBlockHeight\":724922,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "1.1.0-1-g35ba2d3",
				CurrentHighestBlockHash:   "000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0",
				CurrentHighestBlockHeight: 724922,
				MinerID:                   miner.MinerID,
				ResultDescription:         "",
				ReturnResult:              mAPISuccess,
				Timestamp:                 "2022-02-02T12:12:02.6089293Z",
				TxID:                      broadcastExample1TxID,
			},
		}, nil
	} else if miner.Name == minercraft.MinerGorillaPool {
		sig := gorillaPoolSig1
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-02T12:12:02.6089293Z\",\"txid\":\"\",\"returnResult\":\"success\",\"resultDescription\":\"\",\"minerId\":\"03ad780153c47df915b3d2e23af727c68facaca4facd5f155bf5018b979b9aeb83\",\"currentHighestBlockHash\":\"000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0\",\"currentHighestBlockHeight\":724922,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "",
				CurrentHighestBlockHash:   "000000000000000006e6745f6a57a1da8096faf9f71dd59b2bab3f2b0219b7a0",
				CurrentHighestBlockHeight: 724922,
				MinerID:                   miner.MinerID,
				ResultDescription:         "",
				ReturnResult:              mAPISuccess,
				Timestamp:                 "2022-02-02T12:12:02.6089293Z",
				TxID:                      broadcastExample1TxID,
			},
		}, nil
	}

	return nil, errors.New("missing miner response")
}

type minerCraftInMempool struct {
	minerCraftTxOnChain
}

func (m *minerCraftInMempool) SubmitTransaction(_ context.Context, miner *minercraft.Miner,
	_ *minercraft.Transaction,
) (*minercraft.SubmitTransactionResponse, error) {
	if miner.Name == minercraft.MinerTaal {
		sig := "30440220008615778c5b8610c29b12925c8eb479f692ad6de9e62b7e622a3951baf9fbd8022014aaa27698cd3aba4144bfd707f3323e12ac20101d6e44f22eb8ed0856ef341a"
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.4.0\",\"timestamp\":\"2022-02-01T15:19:40.889523Z\",\"txid\":\"683e11d4db8a776e293dc3bfe446edf66cf3b145a6ec13e1f5f1af6bb5855364\",\"returnResult\":\"failure\",\"resultDescription\":\"Missing inputs\",\"minerId\":\"030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e\",\"currentHighestBlockHash\":\"00000000000000000652def5827ad3de6380376f8fc8d3e835503095a761e0d2\",\"currentHighestBlockHeight\":724807,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "1.4.0",
				CurrentHighestBlockHash:   "00000000000000000652def5827ad3de6380376f8fc8d3e835503095a761e0d2",
				CurrentHighestBlockHeight: 724807,
				MinerID:                   miner.MinerID,
				ResultDescription:         "Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T15:19:40.889523Z",
				TxID:                      onChainExample1TxID,
			},
		}, nil
	} else if miner.Name == minercraft.MinerMempool {
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:  "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-01T17:47:52.518Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Missing inputs\",\"minerId\":null,\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
					Encoding: utf8Type,
					MimeType: applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "",
				CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
				CurrentHighestBlockHeight: 724816,
				MinerID:                   miner.MinerID,
				ResultDescription:         "ERROR: Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T17:47:52.518Z",
				TxID:                      "",
			},
		}, nil
	} else if miner.Name == minercraft.MinerMatterpool {
		sig := matterCloudSig1
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.1.0-1-g35ba2d3\",\"timestamp\":\"2022-02-01T17:50:15.130Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Missing inputs\",\"minerId\":\"0253a9b2d017254b91704ba52aad0df5ca32b4fb5cb6b267ada6aefa2bc5833a93\",\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "1.1.0-1-g35ba2d3",
				CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
				CurrentHighestBlockHeight: 724816,
				MinerID:                   miner.MinerID,
				ResultDescription:         "ERROR: Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T17:50:15.130Z",
				TxID:                      "",
			},
		}, nil
	} else if miner.Name == minercraft.MinerGorillaPool {
		sig := gorillaPoolSig1
		pubKey := miner.MinerID
		return &minercraft.SubmitTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     miner,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-01T17:52:04.405Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Missing inputs\",\"minerId\":\"03ad780153c47df915b3d2e23af727c68facaca4facd5f155bf5018b979b9aeb83\",\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Results: &minercraft.SubmissionPayload{
				APIVersion:                "",
				CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
				CurrentHighestBlockHeight: 724816,
				MinerID:                   miner.MinerID,
				ResultDescription:         "ERROR: Missing inputs",
				ReturnResult:              mAPIFailure,
				Timestamp:                 "2022-02-01T17:52:04.405Z",
				TxID:                      "",
			},
		}, nil
	}

	return nil, errors.New("missing miner response")
}

type minerCraftTxNotFound struct {
	MinerCraftBase
}

func (m *minerCraftTxNotFound) SubmitTransaction(_ context.Context, miner *minercraft.Miner,
	_ *minercraft.Transaction,
) (*minercraft.SubmitTransactionResponse, error) {
	return &minercraft.SubmitTransactionResponse{
		JSONEnvelope: minercraft.JSONEnvelope{
			Miner:     miner,
			Validated: true,
			JSONEnvelope: envelope.JSONEnvelope{
				Payload:  "{\"apiVersion\":\"\",\"timestamp\":\"2022-02-01T17:47:52.518Z\",\"txid\":\"\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: Mempool conflict\",\"minerId\":null,\"currentHighestBlockHash\":\"0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1\",\"currentHighestBlockHeight\":724816,\"txSecondMempoolExpiry\":0}",
				Encoding: utf8Type,
				MimeType: applicationJSONType,
			},
		},
		Results: &minercraft.SubmissionPayload{
			APIVersion:                "",
			CurrentHighestBlockHash:   "0000000000000000064c900b1fceb316302426aedb2242852530b5e78144f2c1",
			CurrentHighestBlockHeight: 724816,
			MinerID:                   miner.MinerID,
			ResultDescription:         "ERROR: Mempool conflict",
			ReturnResult:              mAPIFailure,
			Timestamp:                 "2022-02-01T17:47:52.518Z",
			TxID:                      "",
		},
	}, nil
}

func (m *minerCraftTxNotFound) QueryTransaction(_ context.Context, miner *minercraft.Miner,
	_ string, _ ...minercraft.QueryTransactionOptFunc,
) (*minercraft.QueryTransactionResponse, error) {
	if miner.Name == minerTaal.Name {
		sig := "304402201aae61ec65500cf38af48e552c0ea0c62c7937805a99ff6b2dc62bad1a23c183022027a0bb97890f92d41e7b333e8f3dec106aedcd16b782f2f8b46501e104104322"
		pubKey := minerTaal.MinerID
		return &minercraft.QueryTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     minerTaal,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.4.0\",\"timestamp\":\"2022-01-24T01:36:23.0767761Z\",\"txid\":\"918c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b\",\"returnResult\":\"failure\",\"resultDescription\":\"No such mempool or blockchain transaction. Use gettransaction for wallet transactions.\",\"minerId\":\"030d1fe5c1b560efe196ba40540ce9017c20daa9504c4c4cec6184fc702d9f274e\",\"txSecondMempoolExpiry\":0}",
					Signature: &sig,
					PublicKey: &pubKey,
					Encoding:  utf8Type,
					MimeType:  applicationJSONType,
				},
			},
			Query: &minercraft.QueryPayload{
				APIVersion:        "1.4.0",
				Timestamp:         "2022-01-24T01:36:23.0767761Z",
				TxID:              notFoundExample1TxID,
				ReturnResult:      mAPIFailure,
				ResultDescription: "No such mempool or blockchain transaction. Use gettransaction for wallet transactions.",
				MinerID:           minerTaal.MinerID,
			},
		}, nil
	} else if miner.Name == minerMempool.Name {
		return &minercraft.QueryTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     minerMempool,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:  "{\"apiVersion\":\"\",\"timestamp\":\"2022-01-24T01:39:58.066Z\",\"txid\":\"918c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: No such mempool or blockchain transaction. Use gettransaction for wallet transactions.\",\"blockHash\":null,\"blockHeight\":null,\"confirmations\":0,\"minerId\":null,\"txSecondMempoolExpiry\":0}",
					Encoding: utf8Type,
					MimeType: applicationJSONType,
				},
			},
			Query: &minercraft.QueryPayload{
				APIVersion:        "", // NOTE: missing from mempool response
				Timestamp:         "2022-01-24T01:39:58.066Z",
				TxID:              notFoundExample1TxID,
				ReturnResult:      mAPIFailure,
				ResultDescription: "ERROR: No such mempool or blockchain transaction. Use gettransaction for wallet transactions.",
				MinerID:           "", // NOTE: missing from mempool response
			},
		}, nil
	} else if miner.Name == minerGorillaPool.Name {
		sig := "3045022100eaf52c498ee79c7deb7f67ebcdb174b446e3f8e826ef6c9faa3e3365c14008a9022036b9a355574af9576e3f2c124855c81c56164df8713d0615bc0be09c50e103c8"
		pubKey := minerGorillaPool.MinerID
		return &minercraft.QueryTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     minerGorillaPool,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"\",\"timestamp\":\"2022-01-24T01:40:41.136Z\",\"txid\":\"918c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b\",\"returnResult\":\"failure\",\"resultDescription\":\"Mixed results\",\"blockHash\":null,\"blockHeight\":null,\"confirmations\":0,\"minerId\":\"03ad780153c47df915b3d2e23af727c68facaca4facd5f155bf5018b979b9aeb83\",\"txSecondMempoolExpiry\":0}",
					Encoding:  utf8Type,
					Signature: &sig,
					PublicKey: &pubKey,
					MimeType:  applicationJSONType,
				},
			},
			Query: &minercraft.QueryPayload{
				APIVersion:        "",
				Timestamp:         "2022-01-24T01:40:41.136Z",
				TxID:              notFoundExample1TxID,
				ReturnResult:      mAPIFailure,
				ResultDescription: "Mixed results",
				MinerID:           minerGorillaPool.MinerID,
			},
		}, nil
	} else if miner.Name == minerMatterPool.Name {
		sig := "304402200abaf73f5b70f225f52dadc328fd5facf689d8e99ddd731b1c8a17522635c2aa022028c5d040402d8ddd7d64d94f2d7dbcd600ac50e1c503ae40fb314e0435c78b7f"
		pubKey := minerMatterPool.MinerID
		return &minercraft.QueryTransactionResponse{
			JSONEnvelope: minercraft.JSONEnvelope{
				Miner:     minerMatterPool,
				Validated: true,
				JSONEnvelope: envelope.JSONEnvelope{
					Payload:   "{\"apiVersion\":\"1.1.0-1-g35ba2d3\",\"timestamp\":\"2022-01-24T01:41:01.683Z\",\"txid\":\"918c26f8227fa99f1b26f99a19648653a1382fb3b37b03870e9c138894d29b3b\",\"returnResult\":\"failure\",\"resultDescription\":\"ERROR: No such mempool transaction. Use -txindex to enable blockchain transaction queries. Use gettransaction for wallet transactions.\",\"blockHash\":null,\"blockHeight\":null,\"confirmations\":0,\"minerId\":\"0253a9b2d017254b91704ba52aad0df5ca32b4fb5cb6b267ada6aefa2bc5833a93\",\"txSecondMempoolExpiry\":0}",
					Encoding:  utf8Type,
					Signature: &sig,
					PublicKey: &pubKey,
					MimeType:  applicationJSONType,
				},
			},
			Query: &minercraft.QueryPayload{
				APIVersion:        "1.1.0-1-g35ba2d3",
				Timestamp:         "2022-01-24T01:41:01.683Z",
				TxID:              notFoundExample1TxID,
				ReturnResult:      mAPIFailure,
				ResultDescription: "ERROR: No such mempool transaction. Use -txindex to enable blockchain transaction queries. Use gettransaction for wallet transactions.",
				MinerID:           minerMatterPool.MinerID,
			},
		}, nil
	}

	return nil, nil
}

type minerCraftUnreachble struct {
	MinerCraftBase
}

func (m *minerCraftUnreachble) FeeQuote(context.Context, *minercraft.Miner) (*minercraft.FeeQuoteResponse, error) {
	return nil, errors.New("minercraft is unreachable")
}
