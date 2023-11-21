package bux

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hexForProcessedTx = "0100000002cb3553424ffc94b59a60fb358b6cb6dfb694aee894dcd1effc0ed0a9052464e3000000006a4730440220515c3bf93d38fa7cc164746fae4bec8b66c60a82509eb553751afa5971c3e41d0220321517fd5c997ab5f8ef0e59048ce9157de46f92b10d882bf898e62f3ee7343d4121038f1273fcb299405d8d140b4de9a2111ecb39291b2846660ebecd864d13bee575ffffffff624fbcb4e68d162361f456b8b4fef6b9e7943013088b32b6bca7f5ced41ff004010000006a47304402203fb24f6e00a6487cf88a3b39d8454786db63d649142ea76374c2f55990777e6302207fbb903d038cf43e13ffb496a64f36637ec7323e5ac48bb96bdb4a885100abca4121024b003d3cf49a8f48c1fe79b711b1d08e306c42a0ab8da004d97fccc4ced3343affffffff026f000000000000001976a914f232d38cd4c2f87c117af06542b04a7061b6640188aca62a0000000000001976a9146058e52d00e3b94211939f68cc2d9a3fc1e3db0f88ac00000000"

type beefTestCase struct {
	testID               int
	name                 string
	outputValue          uint64
	receiverAddress      string
	ancestors            []*beefTestCaseAncestor
	expectedError        bool
	expectedErrorMessage string
}

type beefTestCaseAncestor struct {
	hex         string
	isMined     bool
	bumpJSON    string
	blockHeight int
}

// TODO: BUX-168 - fix this test
func Test_ToBeef(t *testing.T) {
	testCases := []beefTestCase{
		{
			testID: 1,
			name:   "all parents txs are already mined",
			ancestors: []*beefTestCaseAncestor{
				{
					hex:         "0100000001cfc39e3adcd58ed58cf590079dc61c3eb6ec739abb7d22b592fb969d427f33ee000000006a4730440220253e674e64028459457d55b444f5f3dc15c658425e3184c628016739e4921fd502207c8fe20eb34e55e4115fbd82c23878b4e54f01f6c6ad0811282dd0b1df863b5e41210310a4366fd997127ad972b14c56ca2e18f39ca631ac9e3e4ad3d9827865d0cc70ffffffff0264000000000000001976a914668a92ff9cb5785eb8fc044771837a0818b028b588acdc4e0000000000001976a914b073264927a61cf84327dea77414df6c28b11e5988ac00000000",
					isMined:     true,
					bumpJSON:    `{"blockHeight":"817574","path":[[{"offset":"11432","hash":"3b535e0f8e266124bce9868420052d5a7585c67e82c1edc2c7fe05fd5e140307"},{"offset":"11433","hash":"e3642405a9d00efcefd1dc94e8ae94b6dfb66c8b35fb609ab594fc4f425335cb","txid":true}],[{"offset":"5717","hash":"6ef9c6dde7fff82fa893754109f12378c8453b47dc896596b5531433093ab5b7"}],[{"offset":"2859","hash":"daa67e00ad2aef787998b66cbb3417033fbec136da1e230a5f5df3186f5c0880"}],[{"offset":"1428","hash":"bc777a80d951fbf2b7bd3a8048a9bb78fbf1d23d4127290c3fed9740b4246dd2"}],[{"offset":"715","hash":"762b57f88e7258f5757b48cda96d075cbe767c0a39a83e7109574555fd2dd8ba"}],[{"offset":"356","hash":"bbaab745bcca4f8a4be39c06c7e9be3aa1994f32271e3c6b4f768897153e5522"}],[{"offset":"179","hash":"817694ccbde5dbf88f290c30e8735991708a3d406740f7dd31434ff516a5bfde"}],[{"offset":"88","hash":"ed5b52ba4af9198d398e934a84e18405f49e7abde91cafb6dfe5aeaedb33a979"}],[{"offset":"45","hash":"0e51ec9dd5319ceb32d2d20f620c0ca3e0d918260803c1005d49e686c9b18752"}],[{"offset":"23","hash":"08ab694ef1af4019e2999a543a632cf4a662ae04d5fee879c6aadaeb749f6374"}],[{"offset":"10","hash":"4223f47597b14ee0fa7ade08e611ec80948b5fa9da267ce6c8e5d952e7fdb38e"}],[{"offset":"4","hash":"b6dace0d2294fd6e0c11f74376b7f0a1fc8ee415b350caf90c3ae92749e2a8ee"}],[{"offset":"3","hash":"795e7514ebf6d63b454d3f04854e1e0db0ac3a549f61135d5e9ef8d5785f2c68"}],[{"offset":"0","hash":"3f458f2c06493c31cbc3a035ba131913b274ac7915b9b9bc79128001a75cf76d"}],[{"offset":"1","hash":"b9b9f80cc72a674e37b54a9fdee72a9bff761f8cbcb94146afc2bffef33be89f"}]]}`,
					blockHeight: 817574,
				},
				{
					hex:         "0100000001a114c7deb8deba851d87755aa10aa18c97bd77afee4e1bad01d1c50e07a644eb010000006a473044022041abd4f93bd1db1d0097f2d467ae183801d7842d23d0605fa9568040d245167402201be66c96bef4d6d051304f6df2aecbdfe23a8a05af0908ef2117ab5388d8903c412103c08545a40c819f6e50892e31e792d221b6df6da96ebdba9b6fe39305cc6cc768ffffffff0263040000000000001976a91454097d9d921f9a1f55084a943571d868552e924f88acb22a0000000000001976a914c36b3fca5159231033f3fbdca1cde942096d379f88ac00000000",
					isMined:     true,
					bumpJSON:    `{"blockHeight":"819138","path":[[{"offset":"648","hash":"121bca23ca64d9b925c055f89340802e51f0949ab5edf36ad6dffe050d28400e"},{"offset":"649","hash":"04f01fd4cef5a7bcb6328b08133094e7b9f6feb4b856f46123168de6b4bc4f62","txid":true}],[{"offset":"325","hash":"1e5b72effab8fb56da368f25bab8d8fae7891a5dc70c5da1a8dac4f81e75f990"}],[{"offset":"163","hash":"c97efaa344c57f5e0e46676cbf8629fad9f69a7b1f71d6fda8a1e03f2b546328"}],[{"offset":"80","hash":"5069f334f680952ee9abf37ca5f1cf327e5114920f79ec26b108ae7a491e0b3a"}],[{"offset":"41","hash":"a40cf9eb878b35f853198ebf23ac85061253ffb6e20c4c3eb1ac546b2a376f6d"}],[{"offset":"21","hash":"b5f91b76bf448529368e9421a89e6c756d6b92679ce06479557bf8a5dedb10c3"}],[{"offset":"11","hash":"dbf6acf27c7df7bf4a9de100fd6ad7f73db9b0d659e38235cef4eee26fe367a2"}],[{"offset":"4","hash":"e08fbe8bbdda28ff48478b90c909e4dad7acffc6ff5b3e46f8b4c597d76fc180"}],[{"offset":"3","hash":"0ae8ff1e623834f0624d78703498bb986e1d3d2c5f9d172f05b8e839a09ce0b7"}],[{"offset":"0","hash":"1f55fb14746170226dd929e47fedaac59c65d7b4b9b5502c758bca19a76d5bcf"}],[{"offset":"1","hash":"40ab6623661b0a927bbf231c8a1c0bbf1b64b0b0afb665449cac9ac70e8601dc"}]]}`,
					blockHeight: 819138,
				},
			},
			expectedError:        false,
			receiverAddress:      "1A1PjKqjWMNBzTVdcBru27EV1PHcXWc63W",
			outputValue:          3000,
			expectedErrorMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			ctx, client, deferMe := initSimpleTestCase(t)
			defer deferMe()

			var ancestors []*Transaction
			for _, ancestor := range tc.ancestors {
				ancestors = append(ancestors, addAncestor(ctx, ancestor, client, t))
			}
			newTx := createProcessedTx(ctx, t, client, &tc, ancestors)

			mockGetter := NewMockTransactionGetter()
			mockGetter.Init(ancestors)

			// when
			result, err := ToBeef(ctx, newTx, mockGetter)

			// then
			if tc.expectedError {
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			}

			assert.Equal(t, expectedBeefHex[tc.testID], result)
		})
	}
}

func createProcessedTx(ctx context.Context, t *testing.T, client ClientInterface, testCase *beefTestCase, ancestors []*Transaction) *Transaction {
	draftTx := newDraftTransaction(
		testXPub, &TransactionConfig{
			Inputs: createInputsUsingAncestors(ancestors, client),
			Outputs: []*TransactionOutput{{
				To:       testCase.receiverAddress,
				Satoshis: testCase.outputValue,
			}},
			ChangeNumberOfDestinations: 1,
			Sync: &SyncConfig{
				Broadcast:        true,
				BroadcastInstant: false,
				PaymailP2P:       false,
				SyncOnChain:      false,
			},
		},
		append(client.DefaultModelOptions(), New())...,
	)

	transaction := newTransaction(hexForProcessedTx, append(client.DefaultModelOptions(), New())...)
	transaction.draftTransaction = draftTx
	transaction.DraftID = draftTx.ID

	assert.NotEmpty(t, transaction)

	return transaction
}

func addAncestor(ctx context.Context, testCase *beefTestCaseAncestor, client ClientInterface, t *testing.T) *Transaction {
	grandpaTx := newTransaction(testCase.hex, append(client.DefaultModelOptions(), New())...)

	if testCase.isMined {
		grandpaTx.BlockHeight = uint64(testCase.blockHeight)

		var bump BUMP
		err := json.Unmarshal([]byte(testCase.bumpJSON), &bump)
		require.NoError(t, err)
		grandpaTx.BUMP = bump
	}

	return grandpaTx
}

func createInputsUsingAncestors(ancestors []*Transaction, client ClientInterface) []*TransactionInput {
	var inputs []*TransactionInput

	for _, input := range ancestors {
		inputs = append(inputs, &TransactionInput{Utxo: *newUtxoFromTxID(input.GetID(), 0, append(client.DefaultModelOptions(), New())...)})
	}

	return inputs
}
