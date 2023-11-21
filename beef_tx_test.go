package bux

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hexForProcessedTx = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff17034d800c2f71646c6e6b2f3285a22ac180341a80c10000ffffffff01c63b4125000000001976a9146bacd757abf099715c1251e7a3388eacb64781ca88ac00000000"

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
					hex:         "0100000005d2e6252f996ab1a6fcbe8911e8f865bb719c0e11787397fd818b5bb1ff554a3c000000006a47304402206bfb91eb220f1fccae581abf080ad559bd2a14c7a2616f7e20dcfebb6403c1fa022013038e4d8748b8493e325c7d86cc3b6544433a2a0bb9ed2b1922cc8e78f36c1541210277e25e8e4ab96e46d94a037411d195b537810b1d7301c2d72e9eb40bb47aae34ffffffff827f1758c64b4a0b1226c54316ddeed4618500f026602ebca3cd4b96174a690a000000006a47304402207b6086ca2547a8bf5ec57eb665b085040758a4dc28f560bb6b2ea3c6934df4310220344eadd086a40fe1007c246a3135e907382080dc58c06b9be30abff476e25662412103138a3aac623a5fc9789cd9476efad0605dea61f3e7cd5089195eb0b446b6ab4effffffff0dacf934645c462a155ca35453ab578e7d510687fda689a565e000fc4df11cd5000000006a47304402201f11732e7ed47381ea09af3004e9e1f69a1281cc530f3825fd320fc9e331ff6802200f61b9274f980df180c89f32ff6e30d4676d78000012092d44158cb2a3e8aff9412103ba49c495254796f5ccf0e28f205f62965fafc33367b2b8d6609e5de30c206ad4ffffffff213e4fca3103f812ffcba253caf452c6811947ff6f2fb99b4e18baa1233e84b6010000006b483045022100d676805d8077746d58d79f0199a6f2a0fcf8cc772b149cf77f60ed9b6dd6a60902200c4d805e84f2e0a50d73acffefb3d712a6891f78e9eafddd8bcba872032a88464121035d1d732dbe247c0886753c84dc3d2fc96a9eac26662e8664fe9ce8f67ab6dd98ffffffffe5232220aee5069017d31cc30818bcc971de3e6418f6e62b8cc9a3430d64f3e9010000006b483045022100ed3d4fda64717c43a5fd7329e333d249204e1a25c064bd6bd6b9c945b747befc02205458f90b37aea82c011d4d96c26b808af4b9c19bc07ada196a7a82ac7a3b7eb441210343caa07997898400cefe7a28445b233d30463d13359c1d87ac42ea5da61432a0ffffffff01ce070000000000001976a9141b6b173a4880836d9641b9218ab119a11918684588ac00000000",
					isMined:     true,
					bumpJSON:    `{"blockHeight":"818283","path":[[{"offset":"8922","hash":"6c0eb6ccf943b1941b951ab79b9bdb31b5bdd5ca3e57e517dc07343711164962"},{"offset":"8923","hash":"53b3b48b814bacdab9069b0a857b3ac3613466c4de47e03995d0d4464f41d478","txid":true}],[{"offset":"4460","hash":"d6c5b087486396dc18fffd507a862019ba19b4f78851c7d9e94ff59dac81d6c5"}],[{"offset":"2231","hash":"ecce8f71d6f67037506f84639a18d10fecac3692ff3bfd7dd6f7bdddb39ce1ae"}],[{"offset":"1114","hash":"02ce7190b8d0432fea6199aacbdcd92605ffa3e48b7771628340e3e6195358d8"}],[{"offset":"556","hash":"4c119623561375dfb0305676b5cf82ef9d6e1c15525929734bdbcd84cde3ee6a"}],[{"offset":"279","hash":"70a2e8b11a42d8e94ff2ae6c4d6851d2c7c19ee65c4b033ef31182a4f432cb40"}],[{"offset":"138","hash":"a243839cf6a8a2db84db601879b979a69d62768072a30df0545e7ba3260059ad"}],[{"offset":"68","hash":"0c20f8d761f818a30a9edddc32eb7c3609d6754c279aed73077b01f75f278090"}],[{"offset":"35","hash":"19980db1f3ece269ef131dcddec34c4e8c4e17a2687a872a53f076a83d8298e2"}],[{"offset":"16","hash":"adf3d8e7d3a38a1ec11c49850e00a16d035a71af60124a1d456be8a77e027880"}],[{"offset":"9","hash":"35754104f41f501a6774e0062fdc17331346772b4257e057e156b5e0bd47a146"}],[{"offset":"5","hash":"e78a15a5ead17636c2a974866603794b55dde074f2541855fcc9f9796d7c8799"}],[{"offset":"3","hash":"1689e6d017cf693423c31a94b82aa9650a9ad41e71f63f6b626397a494f4d1ac"}],[{"offset":"0","hash":"7811f73e7a4331d4969155ec43a3be7e17ec4f827e049f210a16f0cc6ac8fef6"}],[{"offset":"1","hash":"67fb20248fe9ce7a6ab724f3f0da9543021d2df413783ade4281d679b80ed145"}]]}`,
					blockHeight: 818283,
				},
				{
					hex:         "010000000469dc4cc3fdb0d34b01e9f7dbae0af803395928f787e35604003f351f4f7f5252000000006b483045022100ec548fb2c6b2a8190bae147f2c41491213e159bef12853cde6d5729b75baf15f022050b8b7490885c6cd444a0c1c0e12d98e24cd2e2bbdba70489cee949af8d11b4c41210391507ea71726708c34141395e08c4a5ea8d6a3e5ab23943ea6faa36174fd4660ffffffffdf96c0a17051236f5d2154d30a0021ecdfa963dc47c0db86efffb90c30195871010000006b483045022100a71e76781a0734ea4d9909cb0e15e7ac63c1a269dcb03b7f0679c784e25aa831022006508da0c72a48a42d67654d4bc002f1fda530cbc9da98d745e94a163ee4c4a3412103f8c774b6331950021b3ecb445b4a5fe046e78def008061963a1b0aa9901e652fffffffff3ef7bba32922a00f3c4893560303c728cf3bbc68ffb9de43ae6b9a040cf039a8000000006a47304402207e7a4aed790fa0f9d25776aeca3aa9eed6796107da41fd15950273589aea838f0220030c3aaf19d34425bc5cb44a488a7ad57fa817a8b8be378b6a28efd40d25913c4121032e2f6e29622309c4dd4daecc16fe6aaa37b848657f2b0a88b9b589525bb27babffffffffe3b23939f7a29999c106a154e8ac555229aa9a039d9a10cb58ec299e2371af99000000006a47304402201b521759f7f42f69915005bf23dd32596026202f6da51418d4a90964e6b924ad02207ac2e3cf7e81b0f5a86ca42d40bcb77a725afc5f694061915e3801b074ff89b5412102657168443ef2b22d747a98131fe686fc46ac0119527b33cbeaafa79670acad92ffffffff0234080000000000001976a9141c052e355d972d64db87b7a207ada864ff1dc05d88ac2e000000000000001976a914cce55c2204a23c80fcdecc6dfb4ddc06bc8f042488ac00000000",
					isMined:     true,
					bumpJSON:    `{"blockHeight":"818153","path":[[{"offset":"8824","hash":"9e4f2f2fcd2bbf840a42c405b7307568e66c9036c2fe870eea01194446359757"},{"offset":"8825","hash":"7f0beb34d1eefbc8abb95b769fc1f5e4d3bfccb6475c3125f28d2630cf23717a","txid":true}],[{"offset":"4413","hash":"c11c03c95f105600bbbf117aefe7a4a2a9f0ae53a6c6369426cd507ac2c99344"}],[{"offset":"2207","hash":"d5b15de14da6f7fe80b4752a799f08f113e3ad9f5bafb07fd6b41e076319a577"}],[{"offset":"1102","hash":"827776d06c72809d9147f106f8deb0f0977194a2faeaac03bc82b182e107dd6f"}],[{"offset":"550","hash":"8a72c6aa5d0c6ed42ebb7646cbb9c509594af519f5ee17d09e4a11ad26c84b3e"}],[{"offset":"274","hash":"b107005b5545dfa3ae6b46137abb34612b34bcb257fb5c1557bd8381427d1120"}],[{"offset":"136","hash":"c065efb45241a3f86e241481e4a2bd0582d7373cb7770b40e6bc0a22c3eee3f5"}],[{"offset":"69","hash":"d26b160753a99c900a446e6f32843985eab2608409d46d8c2db25301c005714d"}],[{"offset":"35","hash":"d8cb12f57b167ae64cfad52e1d3469a9b80bdee1b79fee9b08d65a38b2ac33f6"}],[{"offset":"16","hash":"2a28b28dd855b3b274e5b1686f2d3885f6aa353e0f24e82496d9a849d722eaea"}],[{"offset":"9","hash":"1e00139e44321ac97e74fba4b7c5715c438cdd55d40465c49943c690347a7d9f"}],[{"offset":"5","hash":"ac0c8796cf95797caeef7b2a28d78fd21ff90c5cbf1f46cb6c74f2429a923992"}],[{"offset":"3","hash":"c5412a1f08bd34be173b18beb78c2cf38699d756ddc83e4d4a21dbc60e0e430a"}],[{"offset":"0","hash":"22fe45ce2d40387b999f604111db3db982862511dd413e7ad321d31d9b522b7c"}],[{"offset":"1","hash":"20b70aaa9b39e2f3133317dbe911437679d1d533e4d3bd55164ccf99220d1e5c"}]]}`,
					blockHeight: 818153,
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
