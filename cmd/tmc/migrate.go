package main

import (
	"context"
	"fmt"
	database "go_chain/database"
	"go_chain/node"
	"os"

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

			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			// all transactions
			var txs = []database.Tx{
				database.NewTx("morteza", "morteza", 3, ""),
				database.NewTx("morteza", "morteza", 700, "reward"),
				database.NewTx("morteza", "babayaga", 2000, ""),
				database.NewTx("morteza", "morteza", 100, "reward"),
				database.NewTx("babayaga", "morteza", 1, ""),
				database.NewTx("babayaga", "caesar", 1000, ""),
				database.NewTx("babayaga", "morteza", 50, ""),
				database.NewTx("morteza", "morteza", 600, "reward"),
				database.NewTx("morteza", "morteza", 24700, "reward"),				
			}

			// genesis block, as a pending block
			genesisBlock := node.NewPendingBlock(
				database.Hash{},//parent
				state.NextBlockNumber(), // block number
				database.NewAccount("morteza"),
				txs,// transactions
			)

			// now mine the block
			minedBlock, err := node.Mine(context.Background(), genesisBlock)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// add the mined block to block.db and state.Balances
			_, err = state.AddBlock(minedBlock)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)
	return migrateCmd
}()