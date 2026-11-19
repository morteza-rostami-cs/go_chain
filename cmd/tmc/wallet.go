package main

import (
	// "github.com/ethereum/go-ethereum/accounts/keystore"
	"fmt"
	"go_chain/wallet"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/spf13/cobra"
)

// add --keystore flag to any command
func addKeystoreFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagKeystoreFile, "", "keystore file")
	// add required flags -> --datadir=./
	addDefaultRequiredFlags(cmd)
	//make --keystore flag required 
	_ = cmd.MarkFlagRequired(flagKeystoreFile)
}

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

	// register: pk-print
	walletCmd.AddCommand(walletPrintPrivateKeyCmd())

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

			// create a privKey and return account address
			address, err := wallet.NewKeyStoreAccount(dataDir, password)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// print new created account
			// account_address => derived from => wallet public key
			// private_key => public_key => account address
			fmt.Printf("new account created: %s\n", address.Hex())
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

// a command to read, keystore file, and decrypt the private key and print it
func walletPrintPrivateKeyCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "pk-print",
		Short: "decrypt keystore file and prints the private + public keys.",
		Run: func(cmd *cobra.Command, args []string) {

			// path to keystore file
			ksFile, _ := cmd.Flags().GetString(flagKeystoreFile)
			// get user password
			var password = getPassPhrase("please enter a password to decrypt the keystore file: ", false)

			// read the file
			keyJson, err := os.ReadFile(ksFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			
			// decrypt keystore file, 
			// decrypt private_key using passphrase
			key, err := keystore.DecryptKey(keyJson, password)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			// print private key and public keys
			spew.Dump(key)
		},
	}

	addKeystoreFlag(cmd)
	return cmd
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