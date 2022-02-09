package utils

import (
	"regexp"

	bscript2 "github.com/libsv/go-bt/v2/bscript"
)

// ScriptTypePubKeyHash alias from bscript
const ScriptTypePubKeyHash = bscript2.ScriptTypePubKeyHash

// ScriptTypeNullData alias from bscript
const ScriptTypeNullData = bscript2.ScriptTypeNullData

// ScriptTypeMultiSig alias from bscript
const ScriptTypeMultiSig = bscript2.ScriptTypeMultiSig

// ScriptTypeNonStandard alias from bscript
const ScriptTypeNonStandard = bscript2.ScriptTypeNonStandard

// ScriptHashType is the type for the deprecated script hash
const ScriptHashType = "scripthash"

// ScriptMetanet is the type for a metanet transaction
const ScriptMetanet = "metanet"

// p2PKHRegexp OP_DUP OP_HASH160 [pubkey hash] OP_EQUALVERIFY OP_CHECKSIG
var p2PKHRegexp, _ = regexp.Compile("^76a914.{40}88ac$")

// p2SHRegexp OP_HASH160 [hash] OP_EQUAL
var p2SHRegexp, _ = regexp.Compile("^a914.{40}87$")

// metanetRegexp OP_FALSE OP_RETURN 1635018093
var metanetRegexp, _ = regexp.Compile("^006a046d65746142")

// opReturnRegexp OP_FALSE OP_RETURN
var opReturnRegexp, _ = regexp.Compile("^006a")

// IsP2PKH Check whether the given string is a p2pkh output
func IsP2PKH(lockingScript string) bool {
	return p2PKHRegexp.MatchString(lockingScript)
}

// IsP2SH Check whether the given string is a p2shHex output
func IsP2SH(lockingScript string) bool {
	return p2SHRegexp.MatchString(lockingScript)
}

// IsMetanet Check whether the given string is a metanet output
func IsMetanet(lockingScript string) bool {
	return metanetRegexp.MatchString(lockingScript)
}

// IsOpReturn Check whether the given string is an op_return
func IsOpReturn(lockingScript string) bool {
	return opReturnRegexp.MatchString(lockingScript)
}

// IsMultiSig Check whether the given string is a multi-sig locking script
func IsMultiSig(lockingScript string) bool {
	script, err := bscript2.NewFromHexString(lockingScript)
	if err != nil {
		return false
	}
	return script.IsMultiSigOut()
}

// GetDestinationType Get the type of output script destination
func GetDestinationType(lockingScript string) string {
	if IsP2PKH(lockingScript) {
		return ScriptTypePubKeyHash
	} else if IsMetanet(lockingScript) {
		// metanet is a special op_return - needs to be checked first
		return ScriptMetanet
	} else if IsOpReturn(lockingScript) {
		return ScriptTypeNullData
	} else if IsP2SH(lockingScript) {
		return ScriptHashType
	} else if IsMultiSig(lockingScript) {
		return ScriptTypeMultiSig
	}

	return ScriptTypeNonStandard
}
