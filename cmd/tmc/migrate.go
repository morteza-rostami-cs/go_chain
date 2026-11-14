package main

import (
	"fmt"
	database "go_chain/database"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// a file for seeding some blocks into block.db

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use: "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			// dataDir, _ := cmd.Flags().GetString(flagDataDir)

			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			block0 := database.NewBlock(
				database.Hash{},// hash
				state.NextBlockNumber(), // Header.Number
				0, // nonce
				uint64(time.Now().Unix()),// time
				[]database.Tx{
					database.NewTx("morteza", "morteza", 3, ""),
					database.NewTx("morteza", "morteza", 700, "reward"),
				},
			)

			// apply block txs to state.balances and store in disk, (mempool was removed)
			block0hash, err := state.AddBlock(block0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block1 := database.NewBlock(
				block0hash,
				state.NextBlockNumber(),
				0, // nonce
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("morteza", "babayaga", 2000, ""),
					database.NewTx("morteza", "morteza", 100, "reward"),
					database.NewTx("babayaga", "morteza", 1, ""),
					database.NewTx("babayaga", "caesar", 1000, ""),
					database.NewTx("babayaga", "morteza", 50, ""),
					database.NewTx("morteza", "morteza", 600, "reward"),
				},
			)

			block1hash, err := state.AddBlock(block1)
			if err != nil {
				fmt.Fprintln(os. Stderr, err)
				os.Exit(1)
			}

			block2 := database.NewBlock(
				block1hash,
				state.NextBlockNumber(),
				0, // nonce
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("morteza", "morteza", 24700, "reward"),
				},
			)

			_, err = state.AddBlock(block2)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)
	return migrateCmd
}()