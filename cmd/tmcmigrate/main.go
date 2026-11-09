package main

import (
	"fmt"
	"go_chain/database"
	"os"
	"path/filepath"
	"time"
)

// we seed some transactions and blocks

func main() {
	cwd, _ := os.Getwd()// root dir
	// replay and load the state
	state, err := database.NewStateFromDisk(filepath.Join(cwd, "data"))
	if err != nil { 
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},// genesis block - parent empty hash
		uint64(time.Now().Unix()),// timestamp
		[]database.Tx{
			database.NewTx("morteza", "morteza", 3, ""),
			database.NewTx("morteza", "morteza", 700, "reward"),
		},
	)

	// apply txs to state.Balances AND add block.txs to the mempool
	state.AddBlock(block0)
	// save the txs inside of the mempool inside block.db 
	block0hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0hash, // parent block
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("morteza", "babayaga", 2000, ""),
			database.NewTx("morteza", "morteza", 100, "reward"),
			database.NewTx("babayaga", "morteza", 1, ""),
			database.NewTx("babayaga", "caesar", 1000, ""),
			database.NewTx("babayaga", "morteza", 50, ""),
			database.NewTx("morteza", "morteza", 600, "reward"),
		},
	)
	state.AddBlock(block1)
	state.Persist()
}