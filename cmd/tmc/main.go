package main

import (
	"fmt"
	// "io/fs"
	"go_chain/fs"
	"os"

	"github.com/spf13/cobra"
)

// flag for passing the /data dir in commands
const flagDataDir = "datadir"
// we should pass --port for running each node
const flagPort = "port"

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
	// tmcCmd.AddCommand(txCmd())// tmc tx
	tmcCmd.AddCommand(runCmd())// start node command
	tmcCmd.AddCommand(migrateCmd)

	err := tmcCmd.Execute()// read command and execute the corresponding operation 
	if err != nil {
		// prints to os.Stderr -- error messages
		fmt.Fprintln(os.Stderr, err)
		// 0 means success
		// 1 means error
		os.Exit(1)
	}
}

// add --dataDir flag to any command
func addDefaultRequiredFlags(cmd *cobra.Command){
	cmd.Flags().String(
		flagDataDir,
		"",
		"Absolute path to the node data dir where the DB will/is stored",
	)
	cmd.MarkFlagRequired(flagDataDir)// flag is required
}

func incorrectUsageErr() error {
	// returns error type
	return fmt.Errorf("incorrect usage")
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)
	return fs.ExpandPath(dataDir)
}