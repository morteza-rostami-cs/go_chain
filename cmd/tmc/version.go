package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versioning info constants
const Major = "1"
const Minor = "3"
const Fix = "0"
// description of this release
// const Verbal = "TX Add && Balances List"
// const Verbal = "Immutable Snapshot"
// const Verbal = "Blockchain"
// const Verbal = "Flexible DataDir"
// const Verbal = "HTTP API"
// const Verbal = "Sync"
// const Verbal = "PoW"
// const Verbal = "PoW Reward"
// const Verbal = "Crypto Authentication"
// const Verbal = "Digital Signatures"
// const Verbal = "Decentralized Auth"
const Verbal = "TX Gas"

// a command should be a cobra.Command struct type

var versionCmd = &cobra.Command{
	Use: "version", // tmc version
	Short: "Describes version.",// see this on tmc --help
	Run: func(cmd *cobra.Command, args []string) {
		// print versioning, on: tmc version
		// returns a formatted string
		fmt.Printf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal)
	},
}


