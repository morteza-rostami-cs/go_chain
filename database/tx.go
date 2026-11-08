package database

// account is just a string now
type Account string

// create a new account with a value
func NewAccount(value string) Account {
	return Account(value)
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
}
//methods
func (t Tx) IsReward() bool {
	// check if transaction has a bonus or reward
	// idea is that, you buy 1000 TMC and you get x% bonus, so, supply is added to our chain. not just transferring the existing tokens.
	return t.Data == "reward"
} 

// create new transaction
func NewTx(
	from Account,
	to Account,
	value uint,
	data string,
) Tx {
	// we are passing values in a positional way, into struct
	return Tx{from, to, value, data}
}
