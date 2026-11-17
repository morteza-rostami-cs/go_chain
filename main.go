package main

import (
	// "crypto/sha256"
	"fmt"
	"go_chain/database"
	"os"
)

func main() {

	state, err := database.NewStateFromDisk("./data")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	fmt.Println(state.NextBlockNumber())
	
}