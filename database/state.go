package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type State struct{
	Balances map[Account]uint
	txMempool []Tx
	dbFile *os.File
	// hash of the last block in our chain, is the snapshot/hash of the chain
	latestBlockHash Hash // <= 👈 snapshot
}
//State methods

// get last block hash
func (self *State) LatestBlockHash() Hash {
	return self.latestBlockHash
}


//👺👺 AddBlock() and AddTx() => are used to update state.balances and mempool => after a new block/transactions patch

// 👺👺 applyBlock() and apply() => are used to update state. after loading genesis.json and block.db -- there, we do not need to add transactions to mempool, we are just replaying the old transactions 

// go over block.txs and call self.AddTs() for each tx in a block
// tmcmigrate => calls :: AddBlock => for migrating old tx.db transactions
func (self *State) AddBlock(b Block) error {
	for _, tx := range b.TXs{
		if err := self.AddTx(tx); err != nil{ return err }
	}
	return nil
}

// apply each tx to state.Balances and push to txMempool
// tmc tx add => calls:: AddTx()
func (self *State) AddTx(tx Tx) error {
	// apply to state.Balances
	if err := self.apply(tx); err != nil { return err }

	// push tx to end of mempool slice
	self.txMempool = append(self.txMempool, tx)

	return nil
}

// hash and store the last block
// returns: Hash of the last block
func (self *State) Persist() (Hash, error) {
	// make a block, using state.latestBlockHash = which will be the parent of the new last block
	newLastBlock := NewBlock(
		self.latestBlockHash,
		uint64(time.Now().Unix()), // unix time
		self.txMempool,// store the entire available mempool in our new last block
	)

	// hash the new last block
	blockHash, err := newLastBlock.Hash()
	if err != nil { return Hash{}, nil }

	// we build the wrapper blockFS, for block.db json storage
	blockFs := BlockFS{
		Key: blockHash,
		Value: newLastBlock,
	}

	// struct to json []byte
	blockFsJson, err := json.Marshal(blockFs)
	if err != nil { return Hash{}, err }

	fmt.Printf("persist new last block to disk: \n")
	fmt.Printf("t%s\n", blockFsJson)

	// new line for our blockFS json
	blockFsJson = append(blockFsJson, '\n')

	// write blockFsJson to block.db
	if _, err := self.dbFile.Write(blockFsJson); err != nil{
		return Hash{}, err
	}

	// store the newLastBlock.hash in state, we use this for persisting the next block, and use it as it's parent hash 
	self.latestBlockHash = blockHash
	// empty the mempool, ready for the next patch
	self.txMempool = []Tx{}

	return self.latestBlockHash, nil
}

func (self *State) Close() error {
	return self.dbFile.Close()
}

// apply block.txs to state.Balances
// call this for replay of transactions, when we read from block.db
func (self *State) applyBlock(b Block) error {
	for _, tx := range b.TXs {
		// apply each tx
		if err := self.apply(tx); err != nil { return err }
	}
	return nil
}

// actual apply() logic, reward tx and normal tx
// apply a tx => to state.Balances
func (self *State) apply(tx Tx) error {

	// tx: reward
	if tx.IsReward(){
		self.Balances[tx.To] += tx.Value
		return nil
	}

	if self.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance\n")	
	}

	// tx: normal
	self.Balances[tx.From] -= tx.Value
	self.Balances[tx.To] += tx.Value

	return nil
}

// load genesis.json
// read transactions from block.db
// take root dir path
func NewStateFromDisk(dataDir string) (*State, error) {

	// create /data if does not exist
	// create:: /data /data/genesis.json and /data/block.db
	err := initDataDirIfNotExists(dataDir)
	if err != nil { return nil, err }

	// cwd, err := os.Getwd()// path to root dir
	// if err != nil {return nil, err}// nill for *pointer

	// load the genesis.json
	// var genesisPath = filepath.Join(cwd, "data", "genesis.json")
	gen, err := LoadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil { return nil, err }

	// gen.Balances -> map

	var balances = make(map[Account]uint)// get all balances here

	// get init balances from genesis
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	// open block.db
	// file, err => as long as left side is a new variable, the right err -> just reassigns the previously declared err.
	// var blockDbPath = filepath.Join(cwd, "data", "block.db")
	dbFile, err := os.OpenFile(
		// blockDbPath,// path 
		getBlocksDbFilePath(dataDir),
		// flags:: O_CREATE :: create the file if does not exists.
		os.O_APPEND|os.O_RDWR|os.O_CREATE, 
		0600)

	if err != nil { return nil, err }

	// scanner for reading bytes, one line for each block
	scanner := bufio.NewScanner(dbFile)

	// pointer to empty state
	state := &State{
		Balances: balances,// gen.Balances
		txMempool: make([]Tx, 0),// empty slice, not nil		
		dbFile: dbFile,// pointer to db file
		latestBlockHash: Hash{},// empty hash
	}

	// loop each block line inside block.db
	for scanner.Scan() {
		// check scanning error
		if err := scanner.Err(); err != nil { return nil, err }

		// get json bytes/objects that are => BlockFS 
		/*
		{
		key: 2342323,
		value: {
			header: {},
			payload: [tx, tx],
		}
		}
		*/ 
		blockFsJson := scanner.Bytes()

		var blockFs BlockFS
		// json bytes => *struct
		if err := json.Unmarshal(blockFsJson, &blockFs); err != nil{
			return nil, err
		}

		// apply block => state.Balances map
		// blockFs.Value => Block
		if err := state.applyBlock(blockFs.Value); err != nil {
			return nil, err
		}

		// at the last loop we get the key/hash of the last block, and store it as our snapshot or the whole chain
		state.latestBlockHash = blockFs.Key
	}

	return state, nil
}