package main

import (
	"fmt"
	"go_chain/node"
	"os"

	"github.com/spf13/cobra"
)

// command for running our http node

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use: "run",// tmc run
		Short: "Launches the TMC node and its HTTP api.",
		Run: func(cmd *cobra.Command, args []string){
			// get dataDir flag
			dataDir, _ := cmd.Flags().GetString(flagDataDir)
			fmt.Println("Launching TMC node and its http api.")

			// run our node server
			err := node.Run(dataDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	// register --datadir flag
	addDefaultRequiredFlags(runCmd)
	return runCmd
}