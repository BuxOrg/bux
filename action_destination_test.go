package bux

import (
	"context"
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_NewDestination will test the method NewDestination()
func (ts *EmbeddedDBTestSuite) TestClient_NewDestination() {

	for _, testCase := range dbTestCases {

		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			ctx := context.Background()

			_, err := tc.client.NewXpub(ctx, testXPub, tc.client.DefaultModelOptions()...)
			assert.NoError(t, err)

			metadata := &map[string]interface{}{
				"test-key": "test-value",
			}

			var destination *Destination
			destination, err = tc.client.NewDestination(
				ctx, testXPub, utils.ChainExternal, utils.ScriptTypePubKeyHash, metadata,
			)
			assert.NoError(t, err)
			assert.Equal(t, "fc1e635d98151c6008f29908ee2928c60c745266f9853e945c917b1baa05973e", destination.ID)
			assert.Equal(t, testXPubID, destination.XpubID)
			assert.Equal(t, utils.ScriptTypePubKeyHash, destination.Type)
			assert.Equal(t, utils.ChainExternal, destination.Chain)
			assert.Equal(t, uint32(0), destination.Num)
			assert.Equal(t, testExternalAddress, destination.Address)
			assert.Equal(t, "test-value", destination.Metadata["test-key"])

			destination2, err2 := tc.client.NewDestination(
				ctx, testXPub, utils.ChainExternal, utils.ScriptTypePubKeyHash, metadata,
			)
			assert.NoError(t, err2)
			assert.Equal(t, testXPubID, destination2.XpubID)
			// assert.Equal(t, "1234567", destination2.Metadata[ReferenceIDField])
			assert.Equal(t, utils.ScriptTypePubKeyHash, destination2.Type)
			assert.Equal(t, utils.ChainExternal, destination2.Chain)
			assert.Equal(t, uint32(1), destination2.Num)
			assert.NotEqual(t, testExternalAddress, destination2.Address)
			assert.Equal(t, "test-value", destination2.Metadata["test-key"])
		})

		ts.T().Run(testCase.name+" - error - missing xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			destination, err := tc.client.NewDestination(
				context.Background(), testXPub, utils.ChainExternal,
				utils.ScriptTypePubKeyHash,
				&map[string]interface{}{
					"test-key": "test-value",
				},
			)
			require.Error(t, err)
			require.Nil(t, destination)
			assert.ErrorIs(t, err, ErrMissingXpub)
		})
	}
}

// TestClient_NewDestinationForLockingScript will test the method NewDestinationForLockingScript()
func (ts *EmbeddedDBTestSuite) TestClient_NewDestinationForLockingScript() {

	for _, testCase := range dbTestCases {

		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, err := tc.client.NewXpub(tc.ctx, testXPub, tc.client.DefaultModelOptions()...)
			assert.NoError(t, err)

			lockingScript := "14c91e5cc393bb9d6da3040a7c72b4b569b237e450517901687f517f7c76767601ff9c636d75587f7c6701fe" +
				"9c636d547f7c6701fd9c6375527f7c67686868817f7b6d517f7c7f77605b955f937f517f787f517f787f567f01147f527f7577" +
				"7e777e7b7c7e7b7c7ea77b885279887601447f01207f75776baa517f7c818b7c7e263044022079be667ef9dcbbac55a06295ce" +
				"870b07029bfcdb2dce28d959f2815b16f8179802207c7e01417e2102b405d7f0322a89d0f9f3a98e6f938fdc1c969a8d1382a2" +
				"bf66a71ae74a1e83b0ad046d6574612102b8e6b4441609460d1605ce328d7a39e7216050e105738725b05b7b542dcf1f51205f" +
				"a8b5671a8b577a44ea2d1e70ca9c291145d3da3a7c649fc4e9ea389a8053646c886d76a9146e12c6d84b06757bd4316c33cac4" +
				"4e1e5965589088ac6a0b706172656e74206e6f6465"
			var destination *Destination
			destination, err = tc.client.NewDestinationForLockingScript(
				tc.ctx, testXPub, lockingScript, utils.ScriptTypeNonStandard, map[string]interface{}{"test_key": "test_value"},
			)
			assert.NoError(t, err)
			assert.Equal(t, "a64c7aca7110c7cde92245252a58bb18a4317381fc31fc293f6aafa3fcc7019f", destination.ID)
			assert.Equal(t, testXPubID, destination.XpubID)
			assert.Equal(t, utils.ScriptTypeNonStandard, destination.Type)
			assert.Equal(t, utils.ChainExternal, destination.Chain)
			assert.Equal(t, uint32(0), destination.Num)
			assert.Equal(t, "test_value", destination.Metadata["test_key"])
		})

		ts.T().Run(testCase.name+" - error - missing locking script", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			destination, err := tc.client.NewDestinationForLockingScript(
				tc.ctx, testXPub, "",
				utils.ScriptTypeNonStandard,
				map[string]interface{}{"test_key": "test_value"},
			)
			require.Error(t, err)
			require.Nil(t, destination)
			assert.ErrorIs(t, err, ErrMissingLockingScript)
		})
	}
}

// TestClient_GetDestinations will test the method GetDestinations()
func (ts *EmbeddedDBTestSuite) TestClient_GetDestinations() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, _, rawKey := CreateNewXPub(tc.ctx, t, tc.client)

			// Create a new destination
			destination, err := tc.client.NewDestination(
				tc.ctx, rawKey, utils.ChainExternal, utils.ScriptTypePubKeyHash,
				&map[string]interface{}{
					ReferenceIDField: testReferenceID,
					testMetadataKey:  testMetadataValue,
				},
			)
			require.NoError(t, err)
			require.NotNil(t, destination)

			var getDestinations []*Destination
			getDestinations, err = tc.client.GetDestinations(
				tc.ctx, rawKey, nil,
			)
			require.NoError(t, err)
			require.NotNil(t, getDestinations)
			assert.Equal(t, 1, len(getDestinations))
			assert.Equal(t, destination.Address, getDestinations[0].Address)
			assert.Equal(t, testReferenceID, getDestinations[0].Metadata[ReferenceIDField])
			assert.Equal(t, destination.XpubID, getDestinations[0].XpubID)
		})

		ts.T().Run(testCase.name+" - no destinations found", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, _, rawKey := CreateNewXPub(tc.ctx, t, tc.client)

			// Create a new destination
			destination, err := tc.client.NewDestination(
				tc.ctx, rawKey, utils.ChainExternal, utils.ScriptTypePubKeyHash,
				&map[string]interface{}{testMetadataKey: testMetadataValue},
			)
			require.NoError(t, err)
			require.NotNil(t, destination)

			// use the wrong xpub
			var getDestinations []*Destination
			getDestinations, err = tc.client.GetDestinations(
				tc.ctx, testXPub, nil,
			)
			require.NoError(t, err)
			assert.Equal(t, 0, len(getDestinations))
		})
	}
}

// TestClient_GetDestinationByAddress will test the method GetDestinationByAddress()
func (ts *EmbeddedDBTestSuite) TestClient_GetDestinationByAddress() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, _, rawKey := CreateNewXPub(tc.ctx, t, tc.client)

			// Create a new destination
			destination, err := tc.client.NewDestination(
				tc.ctx, rawKey, utils.ChainExternal, utils.ScriptTypePubKeyHash,
				&map[string]interface{}{
					ReferenceIDField: testReferenceID,
					testMetadataKey:  testMetadataValue,
				},
			)
			require.NoError(t, err)
			require.NotNil(t, destination)

			var getDestination *Destination
			getDestination, err = tc.client.GetDestinationByAddress(
				tc.ctx, rawKey, destination.Address,
			)
			require.NoError(t, err)
			require.NotNil(t, getDestination)
			assert.Equal(t, destination.Address, getDestination.Address)
			assert.Equal(t, testReferenceID, getDestination.Metadata[ReferenceIDField])
			assert.Equal(t, destination.XpubID, getDestination.XpubID)
		})

		ts.T().Run(testCase.name+" - invalid xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, _, rawKey := CreateNewXPub(tc.ctx, t, tc.client)

			// Create a new destination
			destination, err := tc.client.NewDestination(
				tc.ctx, rawKey, utils.ChainExternal, utils.ScriptTypePubKeyHash,
				&map[string]interface{}{testMetadataKey: testMetadataValue},
			)
			require.NoError(t, err)
			require.NotNil(t, destination)

			// use the wrong xpub
			var getDestination *Destination
			getDestination, err = tc.client.GetDestinationByAddress(
				tc.ctx, testXPub, destination.Address,
			)
			require.Error(t, err)
			require.Nil(t, getDestination)
		})
	}
}

// TestClient_GetDestinationByLockingScript will test the method GetDestinationByLockingScript()
func (ts *EmbeddedDBTestSuite) TestClient_GetDestinationByLockingScript() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, _, rawKey := CreateNewXPub(tc.ctx, t, tc.client)

			// Create a new destination
			destination, err := tc.client.NewDestination(
				tc.ctx, rawKey, utils.ChainExternal, utils.ScriptTypePubKeyHash,
				&map[string]interface{}{
					ReferenceIDField: testReferenceID,
					testMetadataKey:  testMetadataValue,
				},
			)
			require.NoError(t, err)
			require.NotNil(t, destination)

			var getDestination *Destination
			getDestination, err = tc.client.GetDestinationByLockingScript(
				tc.ctx, rawKey, destination.LockingScript,
			)
			require.NoError(t, err)
			require.NotNil(t, getDestination)
			assert.Equal(t, destination.Address, getDestination.Address)
			assert.Equal(t, destination.LockingScript, getDestination.LockingScript)
			assert.Equal(t, testReferenceID, getDestination.Metadata[ReferenceIDField])
			assert.Equal(t, destination.XpubID, getDestination.XpubID)
		})

		ts.T().Run(testCase.name+" - invalid xpub", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			_, _, rawKey := CreateNewXPub(tc.ctx, t, tc.client)

			// Create a new destination
			destination, err := tc.client.NewDestination(
				tc.ctx, rawKey, utils.ChainExternal, utils.ScriptTypePubKeyHash,
				&map[string]interface{}{testMetadataKey: testMetadataValue},
			)
			require.NoError(t, err)
			require.NotNil(t, destination)

			// use the wrong xpub
			var getDestination *Destination
			getDestination, err = tc.client.GetDestinationByLockingScript(
				tc.ctx, testXPub, destination.LockingScript,
			)
			require.Error(t, err)
			require.Nil(t, getDestination)
		})
	}
}
