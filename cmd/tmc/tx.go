package main

import (
	"fmt"
	"go_chain/database"
	"os"

	"github.com/spf13/cobra"
)

// command for sending transaction, A to B with x value
//tbb tx add --from Alice --to Bob --value 25

// flag names: --from , --to
const flagFrom = "from"
const flagTo = "to"
const flagValue = "value"

// tmc tx
func txCmd() *cobra.Command {

	// parent command:: tmc tx
	var txCmd = &cobra.Command{
		Use: "tx",
		Short: "Interact with txs (add...).",
		// runs on:: tmc tx -- stop the user from running:: tmc tx only
		PreRunE: func (cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		// parent command has no operation
		Run: func(cmd *cobra.Command, args []string){},
	}

	// register child command:: tmc tx add (adds a new transaction)
	txCmd.AddCommand(txAddCmd())

	return txCmd
}

// add transaction command
func txAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "add",
		Short: "Adds new TX to database.",

		// command operation
		//tbb tx add --from Alice --to Bob --value 25
		Run: func(cmd *cobra.Command, args []string) {
			// get the flags 
			from, _ := cmd.Flags().GetString(flagFrom)
			to, _ := cmd.Flags().GetString(flagTo)
			value, _ := cmd.Flags().GetUint(flagValue)

			// create a transaction
			var tx = database.NewTx(
				database.NewAccount(from),
				database.NewAccount(to),
				value,
				"", // data, reward
			)

			// load the old state, then we add new transaction to mempool
			state, err := database.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			defer state.Close()

			// apply new tx to state.Balances AND add tx to state.txtMempool
			err = state.Add(tx)
			// here transaction can fail, for insufficient funds
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			// save transactions from mempool into disk
			err = state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("TX successfully added to the ledger.")
		},
	}

	// register command flags
	cmd.Flags().String(flagFrom, "", "From what account to send tokens")
	cmd.MarkFlagRequired(flagFrom)// flag is required

	cmd.Flags().String(flagTo, "", "To what account to send tokens")
	cmd.MarkFlagRequired(flagTo)

	cmd.Flags().Uint(flagValue, 0, "How many tokens to send")
	cmd.MarkFlagRequired(flagValue)

	return cmd
}