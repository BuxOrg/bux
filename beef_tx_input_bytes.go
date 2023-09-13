package bux

import (
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
)

var cefExtendedDataMarker = byte(0xEF)
var cefNoExtendedDataMarker = byte(0x00)

func btInputToCefBytes(in *bt.Input, compountedPaths CMPSlice) []byte {
	// get beef bytes
	rawTxInput := in.Bytes(false)
	cefTxInput := btInputGetConditionalExtendedData(in, compountedPaths)

	// compose beef
	buffer := make([]byte, 0, len(rawTxInput)+len(cefTxInput))
	buffer = append(buffer, rawTxInput...)
	buffer = append(buffer, cefTxInput...)

	return buffer
}

func btInputGetCompountedMarklePathIndex(in *bt.Input, compountedPaths CMPSlice) int {
	pathIdx := -1
	previousTxID := string(in.PreviousTxID())

	for i, cmp := range compountedPaths {
		for txID := range cmp[0] {
			if txID == previousTxID {
				pathIdx = i
			}
		}
	}

	return pathIdx
}

func btInputGetConditionalExtendedData(in *bt.Input, compountedPaths CMPSlice) []byte {
	pathIdx := btInputGetCompountedMarklePathIndex(in, compountedPaths)

	if pathIdx > -1 {
		cef := make([]byte, 0)

		cef = append(cef, cefExtendedDataMarker)
		cef = append(cef, getPathIndexBytes(pathIdx)...)
		cef = append(cef, getPreviousTxSatoshisBytes(in)...)
		cef = append(cef, getPreviousTxScriptBytes(in)...)

		return cef

	} else {
		// CEF byte marker - No extended data
		return []byte{cefNoExtendedDataMarker}
	}
}

func getPathIndexBytes(pathIdx int) []byte {
	return bt.VarInt(pathIdx).Bytes()
}

func getPreviousTxSatoshisBytes(in *bt.Input) []byte {
	return utils.LittleEndianBytes64(in.PreviousTxSatoshis, 8)
}

func getPreviousTxScriptBytes(in *bt.Input) []byte {
	if in.PreviousTxScript != nil {
		buffer := make([]byte, 0)

		prevTxScriptLen := uint64(len(*in.PreviousTxScript))
		buffer = append(buffer, bt.VarInt(prevTxScriptLen).Bytes()...)
		buffer = append(buffer, *in.PreviousTxScript...)

		return buffer
	} else {
		return []byte{0x00} // The length of the script is zero
	}

}
