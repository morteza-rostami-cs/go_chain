package main

import (
	"fmt"
	"go_chain/database"
)

func main() {
	fmt.Println("hi, fuckers")

	var account database.Account = database.NewAccount("1")
	account2 := database.NewAccount("2")

	fmt.Println(account == account2)// false

	res, err := database.LoadGenesis("./database/genesis.json")	

	if err != nil{
		fmt.Println(err)
		return
	}

	fmt.Println(res.Balances["morteza"])

	state, err := database.NewStateFromDisk()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(state)
}

