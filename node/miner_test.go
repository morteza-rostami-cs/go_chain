package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"go_chain/database"
	"go_chain/wallet"
	"testing"
	"time"
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

func TestInvalidBlockHash(t *testing.T) {
	hexHash := "000001fa04f8160395c387277f8b2f14837603383d33809a4db586086168edfa"
	var hash = database.Hash{}

	hex.Decode(hash[:], []byte(hexHash))

	isValid := database.IsBlockHashValid(hash)
	// if it is valid, our validation does not work
	if isValid {
		t.Fatal("this hash is not suppose to be valid.")
	}
}

// test the mining algo -- make some block, just for test
func createRandomPendingBlock(miner database.Account) PendingBlock {

	var mortezaAcc = database.NewAccount(wallet.MortezaAccount)
	var babayagaAcc = database.NewAccount(wallet.BabaYagaAccount)

	// this is a genesis block
	return NewPendingBlock(
		database.Hash{},// zeros parent
		0, // block number => last block chain has
		miner,
		[]database.Tx{
			database.NewTx(mortezaAcc, babayagaAcc, 3, ""),
			database.NewTx(mortezaAcc, mortezaAcc, 700, "reward"),
		},
	)
}

// test the mining algo
func TestMine(t *testing.T) {
	// get a pending block
	var miner = database.NewAccount(DefaultMiner)
	var pendingBlock = createRandomPendingBlock(miner) 

	// get a context for stopping the goroutine
	var ctx = context.Background() 

	// call the goroutine Mine()
	fmt.Println("start testing the mining algo")
	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil { t.Fatal(err) }

	// we can rehash our mined block (validator node)
	// notice: mined block at this point has the successful nonce, in its header. => so: it should give us a hash/hex with 6 zeros.
	minedBlockHash, err := minedBlock.Hash()
	if err != nil { t.Fatal(err) }
	
	// fmt.Println(minedBlockHash)
	// check if the hash passes validation rule: to begin with 6 hex zeros
	if !database.IsBlockHashValid(minedBlockHash) {
		t.Fatal() // invalid block sent by miner
	}

	// check if the miner is the same
	if minedBlock.Header.Miner.String() != miner.String() {
		t.Fatal("mined block miner, should be equal miner from pending block.")
	}
}

// test: with passing a context cancellation timeout, into our goroutine
func TestMineWithTimeout(t *testing.T) {
	var miner = database.NewAccount(DefaultMiner)
	pendingBlock := createRandomPendingBlock(miner)

	// our goroutine times out after a certain time
	// we are discarding the cancel function.
	ctx, _ := context.WithTimeout(context.Background(), time.Microsecond * 100)

	_, err := Mine(ctx, pendingBlock)
	if err != nil { t.Fatal(err) }
}

