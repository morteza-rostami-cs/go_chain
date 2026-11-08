package database

import (
	// read file line by line
	// convert json
	// print

	"bufio"
	"encoding/json"
	"fmt"
	"os" // file and operating system
	"path/filepath"
)

//  current state of chain/database

type State struct{
	Balances map[Account]uint
	// excepted transactions that have not been persisted on disk/chain 
	txMempool []Tx // ???
	// pointer to tx.db file
	dbFile *os.File
}
// methods
func (self *State) Add(tx Tx) error {
	// apply transaction to state.balances , update state.Balances(immediately)
	// we apply each transaction to state.balances -- before we push it to mempool
	if err := self.apply(tx); err != nil {
		return err
	}

	// push tx into mempool, to store it on disk later
	self.txMempool = append(self.txMempool, tx)
	return nil
}

// save transactions from txMempool into database/json file
func (self *State) Persist() error {

	// copy the mempool
	mempool := make([]Tx, len(self.txMempool))
	copy(mempool, self.txMempool)

	// go over mempool
	for i :=0; i < len(mempool); i++ {
		// convert transaction into json
		//{"from":"Alice","to":"Bob","value":20}
		txJson, err := json.Marshal(mempool[i])

		if err != nil {return err}

		// append a \n at the end of each transaction, so we have one per file inside our json file
		// txJson is []byte
		txJson = append(txJson, '\n')

		// write json bytes to the disk
		if _, err = self.dbFile.Write( txJson ); err != nil {
			return err // write to file failed
		}

		// remove the persisted tx from state.txMempool
		self.txMempool = self.txMempool[1:] // remove first item

	}
	return nil
}

// close the file, we open this in NewStateFromDisk
func (self *State) Close() { self.dbFile.Close() }

// update state.balances
func (self *State) apply(tx Tx) error {

	// 1. reward transaction
	if tx.IsReward() {// we create money on chain
		// increment the receiver account, without decrementing from the sender. no transfer between accounts.
		self.Balances[tx.To] += tx.Value
		return nil
	}

	// 2. normal transaction
	
	// check if sender has enough fund
	// // uint -> does not give negative -- so it wraps into some big positive -- therefore i use < and not subtraction -
	// if self.Balances[tx.From] - tx.Value < 0 {
	// 	return fmt.Errorf("insufficient balance.")
	// }

	if self.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}

	// subtract from the sender and add to receiver
	self.Balances[tx.From] -= tx.Value	
	self.Balances[tx.To] += tx.Value

	return nil// no error
}

// create a new state, and return a pointer to it
/*
- load initial balances from genesis.json 
- AND open tx.db (keep it open)
- init the chain state
- read lines from tx.db, apply each transaction into state.Balances
- return a pointer to the new state instance
*/
func NewStateFromDisk() (*State, error) {
	
	// get the path to current root directory
	cwd, err := os.Getwd() // /home/apax/__unfuck/go_chain
	if err != nil {
		return nil, err
	}

	// fmt.Println(cwd)

	// load genesis file
	// filepath.Join(), /home/user/myproject/database/genesis.json
	gen, err := LoadGenesis(filepath.Join(cwd, "database", "genesis.json"))
	if err != nil {
		return nil, err
	}

	// gen.Balances -- has the initial state of our database/chain

	// balances for doing our calculations
	var balances = make(map[Account]uint)

	// get the initial balances
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	// open our db file
	// file must exist!!
	file, err := os.OpenFile(
		filepath.Join(cwd, "database", "tx.db"),// path
		// os.O_APPEND :: append to the end of the file
		// os.O_RDWR :: open file for read and write
		// | :: combines these flags 
		os.O_APPEND | os.O_RDWR,
		// permission: only file owner can read and write
		0600,
	)
	if err != nil {
		return nil, err
	}

	// scanner buffer for reading text, line by line
	scanner := bufio.NewScanner(file)// pass file for buffering

	// each line in this file, represent one single transaction

	// create new state
	state := &State{
		Balances: balances,
		// init with a empty slice of tx
		txMempool: make([]Tx, 0),
		// file is open for other parts of program to use
		dbFile: file,
	}

	// read or replay our existing transactions inside tx.db
	for scanner.Scan() {// read until, not line left
		
		// get each transaction
		var tx Tx // struct for each transaction
		// Unmarshal() -- decode json bytes 
		err := json.Unmarshal(
			scanner.Bytes(), // buffered bytes for each line
			&tx, // pointer to tx struct
		)
		if err != nil {return nil, err}

		// apply transaction to state.balances for each account
		if err := state.apply(tx); err != nil {
			return nil, err
		}
	}

	// check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return state, nil
}