package database

import (
	"encoding/json"
	"fmt"

	// "io/ioutil" // deprecated
	"os"
)

// we need to load our genesis file, start data of our chain

// private
type genesis struct{
	// in json: balances: {"morteza": 1000}
	// we only need the balances
	Balances map[Account]uint `json:"balances"`
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