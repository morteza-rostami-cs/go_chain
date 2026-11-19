package node

import (
	"context"
	// "crypto/rand"
	"fmt"
	"go_chain/database"
	"math/rand"
	"time"
)

// candidate block => we add nonce and mine this
type PendingBlock struct{
	parent	database.Hash
	number	uint64
	time	uint64
	// so:each pendingBlock, take its miner into mining process.
	miner	database.Account
	txs	[]database.Tx
}

func NewPendingBlock(
	parent database.Hash,
	number uint64,
	miner	database.Account,
	txs []database.Tx,
) PendingBlock {
	return PendingBlock{
		parent,
		number,
		uint64(time.Now().Unix()), // time
		miner,
		txs,
	}
}

// (goroutine)
// we send a PendingBlock => and spin up bunch of goroutine for mining.
// mining algorithm
// returns a mined block, or error
func Mine(
	ctx context.Context,
	pBlock PendingBlock,
) (database.Block, error) {
	// check for no transactions
	if len(pBlock.txs) == 0 {
		err := fmt.Errorf("mining empty blocks is not allowed")
		return database.Block{}, err
	}

	// state
	var start = time.Now()// start mining time
	var attempt = 0 // num of attempt to find the puzzle/mining
	// hold the current candidate block
	var candidateBlock database.Block 
	// current candidate hash
	var hash *database.Hash = nil // each attempt, a new hash
	// we increment this and try hashing
	var nonce uint32

	// loop as long as the hash is invalid, we start with nil value, cause all zeros Hash array, passes validation test
	// if hash == nil = true => it does not run the second condition, so we do not send nil into IsBlockHashValid(), which expects a Hash.
	for hash == nil || !database.IsBlockHashValid(*hash) {

		// check for goroutine cancellation
		// we need to be able to cancel this for_select loop => in case other nodes mined the same block sooner!
		select {
		case <-ctx.Done():
			// cancel the mining, and return a emptyBlock, error
			fmt.Println("Mining cancelled!")
			err := fmt.Errorf("mining cancelled. %s", ctx.Err())
			return database.Block{}, err
		default:
			// keep mining, so select, does not wait for channel for ever (without default it blocks the loop)
		}

		// increment attempt and generate nonce
		attempt++
		nonce = generateNonce()

		// print info every 1000_000 attempts, to not flood the terminal
		if attempt % 1000_000 == 0 || attempt == 1 {
			fmt.Printf(
				"mining %d Pending txs. attempt: %d\n",
				len(pBlock.txs),
				attempt,
			)
		}

		// make a candidate block
		// if this wins => we return it as a mined block
		candidateBlock = database.NewBlock(
			pBlock.parent,//parent hash
			pBlock.number, // block number
			nonce, // generated nonce in each loop
			pBlock.time,
			pBlock.miner,// block miner
			pBlock.txs,
		)

		// hash the candidate, with the new nonce
		blockHash, err := candidateBlock.Hash()
		if err != nil {
			err = fmt.Errorf("couldn't mine block. %s", err.Error())
			return database.Block{}, nil
		}

		// set the hash, so the next loop uses it
		hash = &blockHash
	}

	// if we get out of this loop, we have mined a block!
	fmt.Printf("\nMined new Block '%x' using PoW:\n", *hash)
	fmt.Printf("\tHeight: '%v'\n", candidateBlock.Header.Number)
	fmt.Printf("\tNonce: '%v'\n", candidateBlock.Header.Nonce)
	fmt.Printf("\tCreated: '%v'\n", candidateBlock.Header.Time)
	fmt.Printf("\tMiner: %v\n", candidateBlock.Header.Miner.String())
	fmt.Printf("\tParent: '%v'\n\n", candidateBlock.Header.Parent.Hex())
	fmt.Printf("\tAttempt: '%v'\n", attempt)
	fmt.Printf("\tTime: %s\n\n", time.Since(start))

	return candidateBlock, nil
}

// generate a random nonce
// nonce: also can be just 0, 1, 2, 3 -- here we are using a random int
// uint32 => has 4.29 billion values, 
// and if nonce is random, nodes try different values, as opposed to all doing: 0, 1, 2, 3
func generateNonce() uint32 {

	// seed it with current time in nano seconds
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Uint32()
}