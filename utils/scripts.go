package utils

import (
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// GetUnlockingScript will generate an unlocking script
func GetUnlockingScript(tx *bt.Tx, inputIndex uint32, privateKey *bec.PrivateKey) (*bscript.Script, error) {
	shf := sighash.AllForkID

	sh, err := tx.CalcInputSignatureHash(inputIndex, shf)
	if err != nil {
		return nil, err
	}

	var sig *bec.Signature
	if sig, err = privateKey.Sign(bt.ReverseBytes(sh)); err != nil {
		return nil, err
	}

	pubKey := privateKey.PubKey().SerialiseCompressed()
	signature := sig.Serialise()

	var s *bscript.Script
	if s, err = bscript.NewP2PKHUnlockingScript(
		pubKey, signature, shf,
	); err != nil {
		return nil, err
	}

	return s, nil
}
