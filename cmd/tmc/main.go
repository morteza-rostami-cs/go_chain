package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// root command line command
	//tmc version
	// tmc command arg1 arg2 ...
	var tmcCmd *cobra.Command = &cobra.Command{
		Use: "tmc", // command name
		Short: "The Morteza Chain CLI",
		// callback cobra runs, when we run a command
		Run: func(
			cmd *cobra.Command,// pointer to command being executed 
			args []string,// cmd args
			){},
	}

	// adding more commands
	// AddCommand takes a cobra.Command type
	tmcCmd.AddCommand(versionCmd)// tmc version
	tmcCmd.AddCommand(balancesCmd()) // tmc balances
	tmcCmd.AddCommand(txCmd())// tmc tx

	err := tmcCmd.Execute()// read command and execute the corresponding operation 
	if err != nil {
		// prints to os.Stderr -- error messages
		fmt.Fprintln(os.Stderr, err)
		// 0 means success
		// 1 means error
		os.Exit(1)
	}
}

func incorrectUsageErr() error {
	// returns error type
	return fmt.Errorf("incorrect usage")
}