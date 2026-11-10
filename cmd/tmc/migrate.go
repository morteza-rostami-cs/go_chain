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
			dataDir, _ := cmd.Flags().GetString(flagDataDir)

			state, err := database.NewStateFromDisk(dataDir)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			var blockNumber = uint64(0)
			if state.LatestBlock() != nil {
				blockNumber = state.LatestBlock().Header.Number + 1
			}

			block0 := database.NewBlock(
				database.Hash{},// hash
				blockNumber, // Header.Number
				uint64(time.Now().Unix()),// time
				[]database.Tx{
					database.NewTx("morteza", "morteza", 3, ""),
					database.NewTx("morteza", "morteza", 700, "reward"),
				},
			)

			state.AddBlock(block0)

			// save block0
			block0hash, err := state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block1 := database.NewBlock(
				block0hash,
				state.LatestBlock().Header.Number + 1,
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

			state.AddBlock(block1)

			block1hash, err := state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block2 := database.NewBlock(
				block1hash,
				state.LatestBlock().Header.Number + 1,
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("morteza", "morteza", 24700, "reward"),
				},
			)
			state.AddBlock(block2)

			_, err = state.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		},
	}

	addDefaultRequiredFlags(migrateCmd)
	return migrateCmd
}()