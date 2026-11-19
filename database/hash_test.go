package database

import (
	"encoding/json"
	"testing"
)

func TestHashUnmarshal(t *testing.T) {
	var h Hash

	err := json.Unmarshal(
		[]byte(`"00000011ba9e00f9512867a114d7fd07735a49dd51f446e2b7d54b415a65b165"`),
		&h,
	)

	if err != nil { t.Fatal(err) }

	t.Log(h.Hex())
}

func TestAccountUnmarshal(t *testing.T) {
	var account Account

	err := json.Unmarshal(
		[]byte(`"0x22ba1F80452E6220c7cc6ea2D1e3EEDDaC5F694A"`),
		&account,
	)

	if err != nil { t.Fatal(err) }

	t.Log(account.Hex())
}

func TestLoadGenesis(t *testing.T){
	gen, err := LoadGenesis("../data/genesis.json") 
	if err != nil { t.Fatal(err) }

	for account, balance := range gen.Balances {
		t.Logf("account: %s, balance: %d", account.Hex(), balance)
	}
}