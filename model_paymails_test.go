package bux

import (
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// todo: refactor, cleanup, test name, add more tests etc

func TestNewPaymail(t *testing.T) {

	t.Run("paymail basic test", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, true, false, WithAutoMigrate(&PaymailAddress{}))
		defer deferMe()

		paymail := "paymail@tester.com"
		xPubID := utils.Hash(testXPub)

		hdKey, err := utils.ValidateXPub(testXPub)
		require.NoError(t, err)
		paymailExternalKey, err := bitcoin.GetHDKeyChild(hdKey, utils.ChainExternal)
		require.NoError(t, err)
		paymailIdentityKey, err := bitcoin.GetHDKeyChild(paymailExternalKey, uint32(utils.MaxInt32))
		require.NoError(t, err)
		paymailExternalXPub := paymailExternalKey.String()
		paymailIdentityXPub := paymailIdentityKey.String()

		p := newPaymail(
			paymail,
			WithClient(client),
		)
		p.Username = "Tester"
		p.Avatar = "img url"
		p.XpubID = xPubID
		p.ExternalXPubKey = paymailExternalXPub
		err = p.Save(ctx)
		require.NoError(t, err)

		p2 := &PaymailAddress{
			Model: *NewBaseModel(ModelPaymail, client.DefaultModelOptions()...),
		}
		conditions := map[string]interface{}{
			"alias":  p.Alias,
			"domain": p.Domain,
		}
		err = Get(ctx, p2, conditions, false, 0)
		require.NoError(t, err)
		identityKey, err := p2.GetIdentityXpub()
		require.NoError(t, err)
		assert.Equal(t, paymail, p2.Alias+"@"+p2.Domain)
		assert.Equal(t, "Tester", p2.Username)
		assert.Equal(t, "img url", p2.Avatar)
		assert.Equal(t, xPubID, p2.XpubID)
		assert.Equal(t, paymailExternalXPub, p2.ExternalXPubKey)
		assert.Equal(t, paymailIdentityXPub, identityKey.String())
		assert.Equal(t, uint32(0), p2.NextIdentityNum)
	})

	t.Run("test derive child keys", func(t *testing.T) {
		// this is used in paymail to store the derived External xpub only in the DB
		hdKey, err := utils.ValidateXPub(testXPub)
		require.NoError(t, err)

		var internal, external string
		external, internal, err = utils.DeriveAddresses(
			hdKey, 1,
		)
		require.NoError(t, err)

		addressExternal := "16fq7PmmXXbFUG5maT5Xvr2zDBUgN1xdMF"
		addressInternal := "1PQW54xMn5KA6uK7wgfzN4y7ZXMi6o7Qtm"
		assert.Equal(t, addressInternal, internal)
		assert.Equal(t, addressExternal, external)

		childKeyChain0, _ := bitcoin.GetHDKeyChild(hdKey, 0)
		childKeyChain01, _ := bitcoin.GetHDKeyChild(childKeyChain0, 1)
		key0, _ := childKeyChain01.ECPubKey()
		address0, _ := bitcoin.GetAddressFromPubKey(key0, true)
		assert.Equal(t, addressExternal, address0.AddressString)

		childKeyChain1, _ := bitcoin.GetHDKeyChild(hdKey, 1)
		childKeyChain11, _ := bitcoin.GetHDKeyChild(childKeyChain1, 1)
		key1, _ := childKeyChain11.ECPubKey()
		address1, _ := bitcoin.GetAddressFromPubKey(key1, true)
		assert.Equal(t, addressInternal, address1.AddressString)
	})
}
