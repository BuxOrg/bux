package utils

import (
	"testing"

	bscript2 "github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
)

var (
	p2pkhHex    = "76a91413473d21dc9e1fb392f05a028b447b165a052d4d88ac"
	p2shHex     = "a9149bc6f9caddaaab28c2bc0a8bf8531f91109bdd5887"
	metanetHex  = "006a046d65746142303237383763323464643466..."
	opReturnHex = "006a067477657463684d9501424945"
	multisigHex = "514104cc71eb30d653c0c3163990c47b976f3fb3f37cccdcbedb169a1dfef58bbfbfaff7d8a473e7e2e6d317b87bafe8bde97e3cf8f065dec022b51d11fcdd0d348ac4410461cbdcc5409fb4b4d42b51d33381354d80e550078cb532a34bfa2fcfdeb7d76519aecc62770f5b0e4ef8551946d8a540911abe3e7854a26f39f58b25c15342af52ae"
	stasHex     = "76a9146d3562a8ec96bcb3b2253fd34f38a556fb66733d88ac6976aa607f5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7c5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01007e818b21414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff007d976e7c5296a06394677768827601249301307c7e23022079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798027e7c7e7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01417e21038ff83d8cf12121491609c4939dc11c4aa35503508fe432dc5a5c1905608b9218ad547f7701207f01207f7701247f517f7801007e8102fd00a063546752687f7801007e817f727e7b01177f777b557a766471567a577a786354807e7e676d68aa880067765158a569765187645294567a5379587a7e7e78637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6867567a6876aa587a7d54807e577a597a5a7a786354807e6f7e7eaa727c7e676d6e7eaa7c687b7eaa587a7d877663516752687c72879b69537a647500687c7b547f77517f7853a0916901247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77788c6301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f777852946301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77686877517f7c52797d8b9f7c53a09b91697c76638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6876638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6863587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f7768587f517f7801007e817602fc00a06302fd00a063546752687f7801007e81727e7b7b687f75537f7c0376a9148801147f775379645579887567726881766968789263556753687a76026c057f7701147f8263517f7c766301007e817f7c6775006877686b537992635379528763547a6b547a6b677c6b567a6b537a7c717c71716868547a587f7c81547a557964936755795187637c686b687c547f7701207f75748c7a7669765880748c7a76567a876457790376a9147e7c7e557967041976a9147c7e0288ac687e7e5579636c766976748c7a9d58807e6c0376a9147e748c7a7e6c7e7e676c766b8263828c007c80517e846864745aa0637c748c7a76697d937b7b58807e56790376a9147e748c7a7e55797e7e6868686c567a5187637500678263828c007c80517e846868647459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e687459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e68687c537a9d547963557958807e041976a91455797e0288ac7e7e68aa87726d77776a14f566909f378788e61108d619e40df2757455d14c010005546f6b656e"
	stas2Hex    = "76a914e130e550626fb267992ea4180f9aaf04ed96357688ac6976aa607f5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7c5f7f7c5e7f7c5d7f7c5c7f7c5b7f7c5a7f7c597f7c587f7c577f7c567f7c557f7c547f7c537f7c527f7c517f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01007e818b21414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff007d976e7c5296a06394677768827601249301307c7e23022079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798027e7c7e7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c8276638c687f7c7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e7e01417e21038ff83d8cf12121491609c4939dc11c4aa35503508fe432dc5a5c1905608b9218ad547f7701207f01207f7701247f517f7801007e8102fd00a063546752687f7801007e817f727e7b01177f777b557a766471567a577a786354807e7e676d68aa880067765158a569765187645294567a5379587a7e7e78637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6878637c8c7c53797e577a7e6867567a6876aa587a7d54807e577a597a5a7a786354807e6f7e7eaa727c7e676d6e7eaa7c687b7eaa587a7d877663516752687c72879b69537a647500687c7b547f77517f7853a0916901247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77788c6301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f777852946301247f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e816854937f77686877517f7c52797d8b9f7c53a09b91697c76638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6876638c7c587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f777c6863587f77517f7c01007e817602fc00a06302fd00a063546752687f7c01007e81687f7768587f517f7801007e817602fc00a06302fd00a063546752687f7801007e81727e7b7b687f75537f7c0376a9148801147f775379645579887567726881766968789263556753687a76026c057f7701147f8263517f7c766301007e817f7c6775006877686b537992635379528763547a6b547a6b677c6b567a6b537a7c717c71716868547a587f7c81547a557964936755795187637c686b687c547f7701207f75748c7a7669765880748c7a76567a876457790376a9147e7c7e557967041976a9147c7e0288ac687e7e5579636c766976748c7a9d58807e6c0376a9147e748c7a7e6c7e7e676c766b8263828c007c80517e846864745aa0637c748c7a76697d937b7b58807e56790376a9147e748c7a7e55797e7e6868686c567a5187637500678263828c007c80517e846868647459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e687459a0637c748c7a76697d937b7b58807e55790376a9147e748c7a7e55797e7e68687c537a9d547963557958807e041976a91455797e0288ac7e7e68aa87726d77776a14f566909f378788e61108d619e40df2757455d14c010005546f6b656e"
)

// TestIsP2PKH will test the method IsP2PKH()
func TestIsP2PKH(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsP2PKH("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsP2PKH(p2pkhHex))
	})

	t.Run("no match - extra data", func(t *testing.T) {
		assert.Equal(t, false, IsP2PKH(p2pkhHex+"06"))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, P2PKHSubstringRegexp.MatchString("somethesetstring"+p2pkhHex+"06rtdhrth"))
	})
}

// TestIsP2SH will test the method IsP2SH()
func TestIsP2SH(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsP2SH("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsP2SH(p2shHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, P2SHSubstringRegexp.MatchString("test"+p2shHex+"test"))
	})
}

// TestIsMetanet will test the method IsOpReturn()
func TestIsMetanet(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsMetanet("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsMetanet(metanetHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, MetanetSubstringRegexp.MatchString("test"+metanetHex+"test"))
	})
}

// TestIsOpReturn will test the method IsOpReturn()
func TestIsOpReturn(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsOpReturn("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsOpReturn(opReturnHex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, OpReturnSubstringRegexp.MatchString("test"+opReturnHex+"test"))
	})
}

// TestIsStas will test the method IsStas()
func TestIsStas(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsStas("nope"))
	})

	t.Run("no match - p2pkhHex", func(t *testing.T) {
		assert.Equal(t, false, IsStas(p2pkhHex))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsStas(stasHex))
	})

	t.Run("match 2", func(t *testing.T) {
		assert.Equal(t, true, IsStas(stas2Hex))
	})

	t.Run("match substring", func(t *testing.T) {
		assert.Equal(t, true, StasSubstringRegexp.MatchString("test"+stas2Hex+"test"))
	})
}

// TestIsMultiSig will test the method IsMultiSig()
func TestIsMultiSig(t *testing.T) {
	t.Parallel()

	t.Run("no match", func(t *testing.T) {
		assert.Equal(t, false, IsMultiSig("nope"))
	})

	t.Run("match", func(t *testing.T) {
		assert.Equal(t, true, IsMultiSig(multisigHex))
	})
}

// TestGetDestinationType will test the method GetDestinationType()
func TestGetDestinationType(t *testing.T) {
	t.Parallel()

	t.Run("no match - non standard", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypeNonStandard, GetDestinationType("nope"))
	})

	t.Run("ScriptTypePubKeyHash", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypePubKeyHash, GetDestinationType(p2pkhHex))
	})

	t.Run("ScriptHashType", func(t *testing.T) {
		assert.Equal(t, ScriptHashType, GetDestinationType(p2shHex))
	})

	t.Run("metanet - ScriptMetanet", func(t *testing.T) {
		assert.Equal(t, ScriptMetanet, GetDestinationType(metanetHex))
	})

	t.Run("op return - ScriptTypeNullData", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypeNullData, GetDestinationType(opReturnHex))
	})

	t.Run("multisig - ScriptTypeMultiSig", func(t *testing.T) {
		assert.Equal(t, bscript2.ScriptTypeMultiSig, GetDestinationType(multisigHex))
	})
}

// TestGetAddressFromScript will test the method GetAddressFromScript()
func TestGetAddressFromScript(t *testing.T) {
	t.Parallel()

	t.Run("p2pkh", func(t *testing.T) {
		assert.Equal(t, "12kwBQPUnAMouxBBWRa5wsA6vC29soEdXT", GetAddressFromScript(p2pkhHex))
	})

	t.Run("stas 1", func(t *testing.T) {
		assert.Equal(t, "1AxScC72W9tyk1Enej6dBsVZNkkgAonk4H", GetAddressFromScript(stasHex))
	})

	t.Run("stas 2", func(t *testing.T) {
		assert.Equal(t, "1MXhcVvUz1LGSkoUFGkANHXkGCtrzFKHpA", GetAddressFromScript(stas2Hex))
	})
}
