package main

import (
	// "crypto/sha256"
	"fmt"
	"go_chain/database"
)

func main() {
	
	var tx = database.Tx{
		From: database.NewAccount("morteza"),
		To: database.NewAccount("babayaga"),
		Value: 1,
		// fixed time, so we can define a nonce, that passes mining rule
		Time: 1579451695,
		Data: "",
	}

	var validSyncedBlock = database.NewBlock(
		database.Hash{},
		1,
		12612040,// nonce
		1579451704,// time
		[]database.Tx{tx},
	)

	// hashed tx
	hash, err := tx.Hash()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(hash.Hex())

	// hash block, that holds that tx
	blockHash, err := validSyncedBlock.Hash()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(blockHash.Hex())
}