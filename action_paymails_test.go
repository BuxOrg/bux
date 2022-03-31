package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var externalXPubID = "xpub69PUyEkuD8cqyA9ekUkp3FwaeW1uyLxbwybEy3bmyD7mM6zShsJqfRCv12B43h6KiEiZgF3BFSMnYLsVZr526n37qsqVXkPKYWQ8En2xbi1"
var testPaymail = "paymail@tester.com"

func (ts *EmbeddedDBTestSuite) TestClient_NewPaymailAddress() {
	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - empty", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false, WithAutoMigrate(&PaymailAddress{}))
			defer tc.Close(tc.ctx)

			paymail := ""
			address, err := tc.client.NewPaymailAddress(tc.ctx, testXPub, paymail, tc.client.DefaultModelOptions()...)
			require.ErrorIs(t, err, ErrMissingPaymailAddress)
			require.Nil(t, address)
		})

		ts.T().Run(testCase.name+" - new paymail address", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false, WithAutoMigrate(&PaymailAddress{}))
			defer tc.Close(tc.ctx)

			paymailAddress, err := tc.client.NewPaymailAddress(tc.ctx, testXPub, testPaymail, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.NotNil(t, paymailAddress)

			assert.Equal(t, "paymail", paymailAddress.Alias)
			assert.Equal(t, "tester.com", paymailAddress.Domain)
			assert.Equal(t, testXPubID, paymailAddress.XpubID)
			assert.Equal(t, externalXPubID, paymailAddress.ExternalXpubKey)

			var p2 *PaymailAddress
			p2, err = getPaymail(tc.ctx, testPaymail, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.NotNil(t, p2)

			assert.Equal(t, "paymail", p2.Alias)
			assert.Equal(t, "tester.com", p2.Domain)
			assert.Equal(t, testXPubID, p2.XpubID)
			assert.Equal(t, externalXPubID, p2.ExternalXpubKey)
		})
	}
}

func (ts *EmbeddedDBTestSuite) Test_DeletePaymailAddress() {
	for _, testCase := range dbTestCases {

		ts.T().Run(testCase.name+" - empty", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false, WithAutoMigrate(&PaymailAddress{}))
			defer tc.Close(tc.ctx)

			paymail := ""
			err := tc.client.DeletePaymailAddress(tc.ctx, paymail, tc.client.DefaultModelOptions()...)
			require.ErrorIs(t, err, ErrMissingPaymail)
		})

		ts.T().Run(testCase.name+" - delete unknown paymail address", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false, WithAutoMigrate(&PaymailAddress{}))
			defer tc.Close(tc.ctx)

			err := tc.client.DeletePaymailAddress(tc.ctx, testPaymail, tc.client.DefaultModelOptions()...)
			require.ErrorIs(t, err, ErrMissingPaymail)
		})

		ts.T().Run(testCase.name+" - delete paymail address", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false, WithAutoMigrate(&PaymailAddress{}))
			defer tc.Close(tc.ctx)

			paymailAddress, err := tc.client.NewPaymailAddress(tc.ctx, testXPub, testPaymail, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.NotNil(t, paymailAddress)

			err = tc.client.DeletePaymailAddress(tc.ctx, testPaymail, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)

			var p2 *PaymailAddress
			p2, err = getPaymail(tc.ctx, testPaymail, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.Nil(t, p2)

			var p3 *PaymailAddress
			p3, err = getPaymailByID(tc.ctx, paymailAddress.ID, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.NotNil(t, p3)
			require.Equal(t, testPaymail, p3.Alias)
			require.True(t, p3.DeletedAt.Valid)
		})
	}
}

// TestClient_UpdateXpubMetadata will test the method UpdatePaymailAddressMetadata()
func (ts *EmbeddedDBTestSuite) TestClient_UpdatePaymailAddressMetadata() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - valid", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)

			metadata := Metadata{
				"test-key-1": "test-value-1",
				"test-key-2": "test-value-2",
				"test-key-3": "test-value-3",
			}
			opts := tc.client.DefaultModelOptions()
			opts = append(opts, WithMetadatas(metadata))

			paymailAddress, err := tc.client.NewPaymailAddress(tc.ctx, testXPub, testPaymail, opts...)
			require.NoError(t, err)
			require.NotNil(t, paymailAddress)

			paymailAddress, err = tc.client.UpdatePaymailAddressMetadata(tc.ctx, testPaymail, Metadata{"test-key-new": "new-value"}, opts...)
			require.NoError(t, err)
			assert.Len(t, paymailAddress.Metadata, 4)
			assert.Equal(t, "new-value", paymailAddress.Metadata["test-key-new"])

			paymailAddress, err = tc.client.UpdatePaymailAddressMetadata(tc.ctx, testPaymail, Metadata{
				"test-key-new-2": "new-value-2",
				"test-key-1":     nil,
				"test-key-2":     nil,
				"test-key-3":     nil,
			}, opts...)
			require.NoError(t, err)
			assert.Len(t, paymailAddress.Metadata, 2)
			assert.Equal(t, "new-value", paymailAddress.Metadata["test-key-new"])
			assert.Equal(t, "new-value-2", paymailAddress.Metadata["test-key-new-2"])

			err = paymailAddress.Save(tc.ctx)
			require.NoError(t, err)

			// todo: make sure it was saved
		})
	}
}
