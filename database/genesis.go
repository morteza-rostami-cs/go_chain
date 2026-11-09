package database

import (
	"encoding/json"
	"fmt"

	// "io/ioutil" // deprecated
	"os"
)

// default data for our genesis.json file
var genesisJson = `
{
  "genesis_time": "2019-03-18T00:00:00.000000000Z",
  "chain_id": "the-morteza-chain-ledger",
  "balances": {
    "morteza": 1000000
  }
}
`

// we need to load our genesis file, start data of our chain

// private
type genesis struct{
	// in json: balances: {"morteza": 1000}
	// we only need the balances
	Balances map[Account]uint `json:"balances"`
}

// write to genesis.json file
func writeGenesisToDisk(path string) error{
	return os.WriteFile(
		path,
		[]byte(genesisJson),
		// os.ModePerm,
		0644,// owner read/write, others: read
		// os.Mode
	)
}

// load genesis from json file
func LoadGenesis(path string) (genesis, error) {
	// read the file
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err); return genesis{}, err		
	}

	// declare this map, to hold: account:balance
	var loadedGenesis genesis
	// load from json
	if err := json.Unmarshal(content, &loadedGenesis); err != nil {
		fmt.Println(err); return genesis{}, err
	}

	return loadedGenesis, nil
}