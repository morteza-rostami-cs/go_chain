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
			// dataDir, _ := cmd.Flags().GetString(flagDataDir)

			port, _ := cmd.Flags().GetUint64(flagPort)
			
			fmt.Println("Launching TMC node and its http api.")

			// create one bootstrap node
			// for now:: all nodes i start, will have 8081 as the initial bootstrap node.
			bootstrap := node.NewPeerNode(
				"127.0.0.1",
				// port + 1,// for now: node.port + 1 
				// so, all nodes have 8081, can connect register them self an pick up other peers.
				node.DefaultHttpPort, // 8081 is the main bootstrap for now.				
				true,// is_bootstrap
				true,// is_active
			)

			// create our node
			n := node.New(getDataDirFromCmd(cmd), port, bootstrap)

			// run our node server
			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	// register --datadir flag
	addDefaultRequiredFlags(runCmd)

	// register port command
	runCmd.Flags().Uint64(
		flagPort,
		node.DefaultHttpPort,// default port
		"exposed HTTP port for communication with peers.",
	)

	return runCmd
}