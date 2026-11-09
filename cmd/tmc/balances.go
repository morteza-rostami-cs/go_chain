package main

import (
	"fmt"
	"go_chain/database"
	"os"

	"github.com/spf13/cobra"
)

// load balances from chain state, and display them

func balancesCmd() *cobra.Command {
	// create a command
	var cmd = &cobra.Command{
		Use: "balances",
		// this command has a child command eg: tmc balances list
		Short: "Interact with balances (list...).",
		// runs before Run , and validates the command usage
		// if: run this: tmc balances -> PreRunE() runs
		// correct command is: tmc balances list
		PreRunE: func(cmd *cobra.Command, arg []string) error {
			return incorrectUsageErr()
		},
		// this runs on parent command: tmc balances, it is empty, cause: we want tmc balances list, not just tmc balances
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// register a child command
	cmd.AddCommand(balancesListCmd())

	return cmd
}


func balancesListCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use: "list",
		Short: "List all balances.",
		// runs on: tmc balances list
		Run: func (cmd *cobra.Command, args []string) {

			// get the --dataDir flag
			dataDir, _ := cmd.Flags().GetString(flagDataDir)

			// replay the blocks in block.db and get the updated state.Balances, pass the dataDir name from command line flag value --flagDir=/data
			state, err := database.NewStateFromDisk(dataDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
	
			// NewStateFromDisk() opens the tx.db file, so if who ever that calls it, also should defer close()		
			defer state.Close()
	
			fmt.Printf("accounts balances at: %x\n", state.LatestBlockHash())		
	
			fmt.Println("Accounts balances:")
			fmt.Println("__________________")
			fmt.Println("")
	
			// state.Balances :: map[Account]uint		
			for account, balance := range state.Balances {
				fmt.Printf("%s: %d\n", account, balance)
			}		
		},
	}

	// setup the --dataDir= flag for list command
	addDefaultRequiredFlags(cmd)

	return cmd
}