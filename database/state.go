package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"reflect"

	// "io/fs"
	"go_chain/fs"
	"os"
)

type State struct{
	Balances	map[Account]uint
	txMempool	[]Tx
	dbFile	*os.File
	// store the whole last block, cause we need it to do: Block.Number + 1, when creating a new block
	latestBlock	*Block

	// hash of the last block in our chain, is the snapshot/hash of the chain
	latestBlockHash	Hash // <= 👈 snapshot
	// does it have 1 genesis block, at least or empty?
	// it's job:: skip number and hash validation on first block.
	hasGenesisBlock	bool
}
//State methods

// get last block hash
func (self *State) LatestBlockHash() Hash {
	return self.latestBlockHash
}

func (self *State) LatestBlock() *Block {
	return self.latestBlock
}

// get current BlockNumber
func (self *State) CurrentBlockNumber() uint64 {
	if self.latestBlock == nil { return uint64(0) }
	return self.latestBlock.Header.Number
}

// what should be the next block number allowed
func (self *State) NextBlockNumber() uint64 {
	// if no block (chain empty): 0
	// if block: blockNum + 1
	if !self.hasGenesisBlock { return uint64(0) }
	return self.latestBlock.Header.Number + 1
}

// our chain is 0 based, so: 0, 1, 2, 3
// return the length of chain
func (self *State) GetChainLen() uint64 {
	// no block on chain yet
	if self.latestBlock == nil { return uint64(0) }	
	// zero based length is always 1 behind
	return self.latestBlock.Header.Number + 1
}

// apply a path of blocks, txs => to balances and mempool
func (self *State) AddBlocks(blocks []Block) error {
	for _, b := range blocks {
		_, err := self.AddBlock(b)
		if err != nil { return err }
	}
	return nil
}

//👺👺 AddBlock() and AddTx() => are used to update state.balances and mempool => after a new block/transactions patch

// 👺👺 applyBlock() and apply() => are used to update state. after loading genesis.json and block.db -- there, we do not need to add transactions to mempool, we are just replaying the old transactions 

// go over block.txs and call self.AddTs() for each tx in a block
// tmcmigrate => calls :: AddBlock => for migrating old tx.db transactions

// apply block to state and persists on disk.
// returns hash, for next AddBlock() call !!
func (self *State) AddBlock(b Block) (Hash, error) {

	// copy state, we change the pendingState, in case one the operation fails, the main state does not change.
	// reason: in applyBlock() we validate block, so it might fail!! and we do not want to change to State til the end.
	pendingState := self.copy()

	// apply block to state
	if err := applyBlock(b, pendingState); err != nil{
		return Hash{}, err
	}

	// hash block
	blockHash, err := b.Hash()
	if err != nil { return Hash{}, nil }

	//blockFs json storage
	blockFs := BlockFS{blockHash, b}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil { return Hash{}, nil }

	fmt.Printf("Persisting new block to disk.")
	fmt.Printf("\t%s\n", blockFsJson[len(blockFsJson)-100:])

	// write blockFs to the block.db file
	blockFsJson = append(blockFsJson, '\n')
	_, err = self.dbFile.Write(blockFsJson)
	if err != nil { return Hash{}, err }

	// set the state to pending state
	// if: blocks successfully persist => set the main state
	self.Balances = pendingState.Balances
	self.latestBlockHash = blockHash

	lastBlock := b // copy block
	self.latestBlock = &lastBlock // store a pointer
	self.hasGenesisBlock = true

	return blockHash, nil
}

// copy state
func (self *State) copy() State {
	s := State{}

	s.hasGenesisBlock = self.hasGenesisBlock
	s.latestBlock = self.latestBlock// not copy, is a pointer
	s.latestBlockHash = self.latestBlockHash
	// s.txMempool = make([]Tx, len(self.txMempool))
	s.txMempool = []Tx{} // empty slice
	s.Balances = make(map[Account]uint)

	//copy balances
	for acc, balance := range self.Balances{
		s.Balances[acc] = balance
	}

	// copy transactions
	for _, tx := range self.txMempool{
		s.txMempool = append(s.txMempool, tx)
	}

	return s
}

// validate and apply a block to state
// take a copy of main state
func applyBlock(b Block, s State) error {
	// // next block number ,we expect
	var nextExpectedBlockNumber = s.NextBlockNumber()

	// if block number does not match
	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf(
			"next expected block must be '%d' not '%d'",
			nextExpectedBlockNumber,
			b.Header.Number,
		)
	}

	// validate parent hash
	// s.latestBlock.Header.Number > 0 => means: first block does not have a parent to check!!
	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf(
			"next block parent hash must be '%x' not '%x'",
			s.latestBlockHash,
			b.Header.Number,
		)
	}

	// if all validation passed => apply transactions to state now.
	return applyTXs(b.TXs, &s)
}

// loop and apply transactions
func applyTXs(txs []Tx, s *State) error {
	for _, tx := range txs {
		if err := applyTx(tx, s); err != nil { return err }
	}
	return nil
}

// apply each transaction to state.Balances
func applyTx(tx Tx, state *State) error {
	// reward tx
	if tx.IsReward() {
		state.Balances[tx.To] += tx.Value
		return nil
	}

	// check for insufficient balance
	if tx.Value > state.Balances[tx.From] {
		fmt.Println("insufficient balance")
		return fmt.Errorf(
			"wrong TX. sender '%s' balance is %d TMC. Tx cost is %d TMC.",
			tx.From,
			state.Balances[tx.From],
			tx.Value,
		)
	}

	// normal tx
	state.Balances[tx.From] -= tx.Value
	state.Balances[tx.To] += tx.Value

	return nil
}

func (self *State) Close() error {
	return self.dbFile.Close()
}

// load genesis.json
// read transactions from block.db
// take root dir path
func NewStateFromDisk(dataDir string) (*State, error) {

	dataDir = fs.ExpandPath(dataDir)// handle ~/

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

		// book:: wants this
		// latestBlock: Block{},
		// nil:: so i can check the genesis block in Persist
		latestBlock: nil,
		latestBlockHash: Hash{},// empty hash
		hasGenesisBlock: false,
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

		// skip empty lines
		if len(blockFsJson) == 0 { break }

		var blockFs BlockFS
		// json bytes => *struct
		if err := json.Unmarshal(blockFsJson, &blockFs); err != nil{
			return nil, err
		}

		// apply block => state.Balances map
		// blockFs.Value => Block
		// if err := state.applyBlock(blockFs.Value); err != nil {
		// 	return nil, err
		// }

		if err := applyTXs(blockFs.Value.TXs, state); err != nil {
			return nil, err
		}

		// book wants this::
		//state.latestBlock = blockFs.Value

		block := blockFs.Value

		state.latestBlock = &block// set last block in state

		// at the last loop we get the key/hash of the last block, and store it as our snapshot or the whole chain
		state.latestBlockHash = blockFs.Key

		// if block.db empty => for loop does not even run, no block!
		state.hasGenesisBlock = true
	}

	return state, nil
}