package main

import (
	"context"
	"fmt"
	database "go_chain/database"
	"go_chain/node"
	"time"

	"github.com/spf13/cobra"
)

// a file for seeding some blocks into block.db

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use: "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("\n == migration command start == \n")
			// dataDir, _ := cmd.Flags().GetString(flagDataDir)
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)
			miner, _ := cmd.Flags().GetString(flagMiner)

			// peer node
			var peer = node.NewPeerNode(
				node.DefaultIP, //ip
				node.DefaultHttpPort, // port
				true, // isBootstrap
				database.NewAccount("morteza"), // account
				false, // connected
			)

			// create a node and let it mine
			var node = node.New(
				getDataDirFromCmd(cmd), //dataDir
				ip,// ip
				port,// port
				database.NewAccount(miner),// acc
				peer,// bootstrap
			)

			// add transactions to node pool
			node.AddPendingTX(database.NewTx("morteza", "morteza", 3, ""), peer)
			node.AddPendingTX(database.NewTx("morteza", "babayaga", 2000, ""), peer)
			node.AddPendingTX(database.NewTx("babayaga", "morteza", 1, ""), peer)
			node.AddPendingTX(database.NewTx("babayaga", "caesar", 1000, ""), peer)
			node.AddPendingTX(database.NewTx("babayaga", "morteza", 50, ""), peer)

			// node timeout and shutdown
			ctx, closeNode := context.WithTimeout(context.Background(), time.Minute * 15)

			// check for mined block and shutdown node server
			go func() {
				ticker := time.NewTicker(time.Second * 10)

				for range ticker.C {
					// if one block at least mined => stop
					if !node.LatestBlockHash().IsEmpty() {
						closeNode()
						return
					}					
				}
			}()

			err := node.Run(ctx)
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	migrateCmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block rewards.")
	migrateCmd.Flags().String(flagIP, node.DefaultIP, "ip for this peer")
	migrateCmd.Flags().Uint64(flagPort, node.DefaultHttpPort, "port for this peer")

	return migrateCmd
}()