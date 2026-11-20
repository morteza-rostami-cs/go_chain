package database

import (
	"crypto/sha256"
	"encoding/json"
	"go_chain/config"

	// "go_chain/wallet"

	// "go_chain/wallet"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

/*
Tx.Hash() / SignedTx.Hash(): SHA-256, used for identity (map keys, block contents).

wallet.Sign / wallet.Verify: the signing hash

when it comes to hashing tx for private key signing also => book eventually uses sha-256 hashing!!

*/

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

// signed tx, tx after signing
// by embedding Tx => we can pass SignedTx , where Tx expected.
type SignedTx struct{
	Tx// embedding, compose Tx methods
	// 65 byte slice
	Sig	[]byte	`json:"signature"`
}
// methods
func (self SignedTx) Hash() (Hash, error) {
	// SignedTx, has Tx.Encode() method, (composition), it returns json for the underlying Tx, => SignedTx.Hash() == SignedTx.Tx.Hash() -- we don't want the sig, in this hash, just Tx!!
	txJson, err := self.Encode()
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(txJson), nil
}

// tx + tx.sig = public key
// public key to account
// check: account == tx.From
// so: sig gives the same account and tx.From => and if we change anything about the tx, sig is not valid.
func (self SignedTx) IsAuthentic() (bool, error) {
	// txHash, err := self.Hash()
	txHash, err := self.Tx.Hash()
	if err != nil { return false, err }

	// tx + tx.sig = public key
	// txHash[:] => array to slice
	recoveredBubKey, err := crypto.SigToPub(txHash[:], self.Sig)
	if err != nil { return false, err }

	//get account from public key 
	var acc = config.GetAccountFromPubKey(*recoveredBubKey)

	// check: publicKey.account == tx.From
	// basically: is the sender also owner of private_key
	return acc.Hex() == self.From.Hex(), nil
}

func NewSignedTx(tx Tx, sig []byte) SignedTx {
	return SignedTx{tx, sig}
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

// so: struct to json bytes => so some other hashing function can hash the same bytes => as inside Hash()
func (self Tx) Encode() ([]byte, error) {
	return json.Marshal(self)
}

// hash transaction
// tx hash will be used as the key => in pendingTxs map
func (self Tx) Hash() (Hash, error) {
	txJson, err := self.Encode()
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