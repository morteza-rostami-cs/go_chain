package main

import (
	// "github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"go_chain/wallet"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/spf13/cobra"
)

//tbb wallet new-account --datadir=$HOME/.c12_crypto
func walletCmd() *cobra.Command {
	var walletCmd = &cobra.Command{
		Use: "wallet",
		Short: "Manages blockchain accounts and keys.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// register: new-account
	walletCmd.AddCommand(walletNewAccountCmd())

	return walletCmd
}

// sub-command: tmc wallet new-account
func walletNewAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "new-account",
		Short: "creates a new account with a new set of elliptic-curve private + public keys.",
		Run: func(cmd *cobra.Command, args []string){
			// account creation
			
			// get a password
			var password = getPassPhrase(
				"Please enter a password to encrypt the new wallet.",
				true,// password confirmation
			)

			// store keys here
			var dataDir = getDataDirFromCmd(cmd)

			// keyphrase encryption settings
			var ks = keystore.NewKeyStore(
				wallet.GetKeystoreDirPath(dataDir),// store encrypted keys here
				keystore.StandardScryptN,
				keystore.StandardScryptP,
			)

			// create an account
			// generate private_key, derives public_key, create eth style account_address, encrypt the private_key using the password we entered, saves the encrypted account data and keys to the keystore directory
			acc, err := ks.NewAccount(password)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// print new created account
			// account_address => derived from => wallet public key
			// private_key => public_key => account address
			fmt.Printf("new account created: %s\n", acc.Address.Hex())
		},
	}

	addDefaultRequiredFlags(cmd)
	return cmd
}

// get wallet password => from cmd
// this passphrase => protects private key on the disk.
func getPassPhrase(
	userPrompt	string,
	confirmation	bool,// should be entered twice
) string {
	// get user passphrase

	password, err := prompt.Stdin.PromptPassword(userPrompt)
	if err != nil {
		utils.Fatalf("failed to read a password: %v", err)
	}

	// basically:: enter password again
	if confirmation {
		confirm, err := prompt.Stdin.PromptPassword("Repeat")
		if err != nil {
			utils.Fatalf("failed to read password confirmation: %v", err)
		}

		// check if password is confirmed
		if password != confirm {
			utils.Fatalf("passwords do not match")
		}
	}
	return password
}

/*

# public key is large!

# so, account_address => is derived from it.
0x22ba1F80452E6220c7cc6ea2D1e3EEDDaC5F694A

0x => means hex

# formula from private_key to account_address

Generate the private key.

Derive the public key using the secp256k1 elliptic curve.

Remove the public key's uncompressed-format prefix.

Calculate Keccak-256 over the remaining public-key bytes.

Take the final 20 bytes of the hash.

Encode those bytes as hexadecimal, with a 0x prefix.
*/