package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versioning info constants
const Major = "0"
const Minor = "1"
const Fix = "0"
// description of this release
const Verbal = "TX Add && Balances List"

// a command should be a cobra.Command struct type

var versionCmd = &cobra.Command{
	Use: "version", // tmc version
	Short: "Describes version.",// see this on tmc --help
	Run: func(cmd *cobra.Command, args []string) {
		// print versioning, on: tmc version
		// returns a formatted string
		fmt.Println(
			fmt.Sprintf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal))
	},
}


