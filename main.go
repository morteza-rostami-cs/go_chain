package main

import (
	// "crypto/sha256"
	"fmt"
	"go_chain/database"

	"github.com/ethereum/go-ethereum/common"
)

func main() {

	// non hex string => gives an empty address
	// 0x0000000000000000000000000000000000000000
	var acc = database.NewAccount("morteza")
	fmt.Println(acc)

	// real 20byte hex string eth address
	// this is hex string
	var ethAddress = "0x8408427314a7ff1a06e15065a77344099269c304"

	var morteza any = database.NewAccount(ethAddress)

	// 0x8408427314a7FF1A06e15065A77344099269c304
	fmt.Println(morteza)

	// morteza => is type Account = common.Address
	if t, ok := morteza.(common.Address); ok {
		fmt.Println(t)
	}
}