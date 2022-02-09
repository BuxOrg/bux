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
