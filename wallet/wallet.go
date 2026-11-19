package wallet

import (
	"crypto/ecdsa"
	"fmt"

	// "go_chain/database"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// sub dir name, for storing encrypted keys
const keystoreDirName = "keystore"

// some default wallet addresses (for testing and so on)
const MortezaAccount = "0x22ba1F80452E6220c7cc6ea2D1e3EEDDaC5F694A"
const BabaYagaAccount = "0x21973d33e048f5ce006fd7b41f51725c30e4b76b"
const CaesarAccount = "0x84470a31D271ea400f34e7A697F36bE0e866a716"

// keys storage path
func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}

// create new private_key and store it in keystore dir
func NewKeyStoreAccount(dataDir string, password string) (common.Address, error) {
	// config
	var ks = keystore.NewKeyStore(
		GetKeystoreDirPath(dataDir),
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)

	// create private_key sign it with passphrase (Address), save to disk
	acc, err := ks.NewAccount(password)
	if err != nil {
		return common.Address{}, err
	}
	// return an eth account, 0x23423...
	return acc.Address, nil
}

// take: a message []byte, and, private key
// return a [65]byte signature
func Sign(
	msg []byte,
	privKey *ecdsa.PrivateKey,
) ([]byte, error) {

	// hash the message or tx or whatever bytes, we need the hash for signing
	// returns: [32]byte keccak-256 hash
	msgHash := crypto.Keccak256(msg)

	// sign the message hash
	sig, err := crypto.Sign(msgHash, privKey)
	if err != nil { return nil, err }

	// check sig length
	if len(sig) != crypto.SignatureLength {
		return nil, fmt.Errorf(
			"wrong size for sig! got %d, want %d",
			len(sig),
			crypto.SignatureLength,
		)
	}

	return sig, nil
}

/*
verify signature:

# this is why we do not need to send the public_key with tx to the node => we just get it from:: msg + sig

message + signature → publicKey

publicKey = Verify(message, signature)
*/

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	// hash the message
	var msgHash = crypto.Keccak256(msg)

	// verify private_key signing, msg + sig = public_key
	publicKey, err := crypto.SigToPub(msgHash, sig)

	// if: msg + sig = invalid! => we get error, (no public_key)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to verify the message sig: %w",
			err,
		)
	}

	return publicKey, nil
}

/*
now, we can convert the public_key to:: eth_address
Public Key
    ↓
Keccak256(Public Key)
    ↓
Last 20 bytes
    ↓
Ethereum Address
*/
// public_key => account_address

