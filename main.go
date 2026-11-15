package main

import (
	// "crypto/sha256"
	"fmt"
	"go_chain/database"
)

func main() {
	// [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
	// we get all zeros => array of [32]byte
	var hash = database.Hash{}
	var hash2 database.Hash

	if hash == hash2 {
		fmt.Println("both are all zeros.")
	}

	fmt.Println(hash)

	// if you want nil, you have to set a pointer to it
	var pHash *database.Hash = nil

	fmt.Println(pHash)
	pHash = &database.Hash{}
	fmt.Println(pHash)

}