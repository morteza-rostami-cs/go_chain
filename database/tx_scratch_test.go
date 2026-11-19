package database

import (
	"fmt"
	"go_chain/wallet"

	// "go_chain/wallet"
	"testing"
)

func TestTxHashStableAndDistinct(t *testing.T) {
	tx1 := NewTx(
		NewAccount(wallet.MortezaAccount),
		NewAccount(wallet.BabaYagaAccount),
		5,
		"",
	)

	// hash transaction
	tx1Hash, err := tx1.Hash()
	if err != nil { t.Fatal(err) }

	// hash it again
	tx1Hash2, _ := tx1.Hash()

	// tx1 must produce the same hash each time
	if tx1Hash != tx1Hash2 {
		t.Fatal("same tx must produce the same hash")
	}

	// a tx created at a different time
	tx2 := NewTx(
		NewAccount(wallet.MortezaAccount),
		NewAccount(wallet.BabaYagaAccount),
		5,
		"",
	)

	tx2Hash, _ := tx2.Hash()

	// transaction created in different times must produce the different hash, even if the content is identical
	fmt.Println(tx1.Time)
	fmt.Println(tx2.Time)
	
	if tx1Hash == tx2Hash {
		fmt.Println(tx1Hash.Hex())
		fmt.Println(tx2Hash.Hex())
		t.Fatal("txs created at different times must differ in hash.")
	}
}