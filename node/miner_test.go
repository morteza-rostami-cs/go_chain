package node

import (
	"encoding/hex"
	"go_chain/database"
	"testing"
)

func TestValidBlockHash(t *testing.T) {

	// some random [32]byte hash in hext
	var hexHash = "000000a3f8c2e1b7d9046f5a8c3e2b1d7f9a0c4e6b8d2f1a3c5e7b9d0f2a4c6e"

	var hash = database.Hash{}

	// convert hex into => raw [32]byte hash
	// hash[:] => gives a slice, with reference to underlying array, we store hash there
	_, err := hex.Decode(hash[:], []byte(hexHash))
	if err != nil {
		t.Fatalf("hash %s is invalid. for some reason.", hexHash)
	}

	// validate the mined hash
	var isValid = database.IsBlockHashValid(hash)
	if !isValid {
		t.Fatalf("hash %s is invalid. we need 6 zeros.", hexHash)
	}
}