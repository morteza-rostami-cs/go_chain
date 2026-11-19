package database

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// account is just a string now
// type Account string

type Account = common.Address // alias
// eth 20byte address
// Account is a distinct type, with underlying type common.Address (meaning: it needs conversion)
// type Account common.Address

// create a new account with a value
// func NewAccount(value string) Account {
// 	return Account(value)
// }

// book does this => use: common.Address => instead of => Account (i keep an alias Account)
func NewAccount(value string) Account {
	// turn a hex string to 20byte address
	return common.HexToAddress(value)
}

// Transaction
type Tx struct{
	// we need tags for json object names
	From Account `json:"from"`
	To Account `json:"to"`
	// token in transfer
	Value uint `json:"value"`
	// meta data
	Data string `json:"data"`

	// add timestamp to tx, so it's hash will be a unique identifier, so we can differentiate between txs, when archiving and so on (for:: transaction_hash)
	Time uint64	`json:"time"`
}
//methods
func (self Tx) IsReward() bool {
	// check if transaction has a bonus or reward
	// idea is that, you buy 1000 TMC and you get x% bonus, so, supply is added to our chain. not just transferring the existing tokens.
	return self.Data == "reward"
} 

// hash transaction
// tx hash will be used as the key => in pendingTxs map
func (self Tx) Hash() (Hash, error) {
	txJson, err := json.Marshal(self)
	if err != nil { return Hash{}, err }

	// hash json bytes of our tx
	return sha256.Sum256(txJson), nil
}

// create new transaction
func NewTx(
	from Account,
	to Account,
	value uint,
	data string,
) Tx {
	// we are passing values in a positional way, into struct
	return Tx{
		From: from, 
		To: to, 
		Value: value, 
		Data: data,
		// this does not capture small time difference
		// Time: uint64(time.Now().Unix()),

		// we will use nano seconds
		Time: uint64(time.Now().UnixNano()),
	}
}
