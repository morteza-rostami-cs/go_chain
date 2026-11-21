package config

import (
	// "crypto"
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// some default wallet addresses (for testing and so on)
const MortezaAccount = "0x22ba1F80452E6220c7cc6ea2D1e3EEDDaC5F694A"
const BabaYagaAccount = "0x21973d33e048f5ce006fd7b41f51725c30e4b76b"
const CaesarAccount = "0x84470a31D271ea400f34e7A697F36bE0e866a716"

// var GetAccountFromPubKey = wallet.GetAccountFromPubKey
func GetAccountFromPubKey(pubKey ecdsa.PublicKey) common.Address {
	// convert to [65]byte
	// for getting eth address we use this => keccak256(pubkey)[12:] => even though we changed it for signing to sha256!!
	var pubKeyBytes = elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	// skip the prefix 0x04 
	var pubKeyBytesHash = crypto.Keccak256(pubKeyBytes[1:])
	// account address => is the last 20 bytes -- take bytes 12..32 (the last 20)
	// var account = common.BytesToAddress(pubKeyBytesHash[:12])
	var account = common.BytesToAddress(pubKeyBytesHash[12:])

	return account
}

// basically: we want to have control over the eth_addresses stored inside keystore files
// fixed account_address => that we store in test keystore file
const TestKsMortezaAccount = "0x3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"

const TestKsBabaYagaAccount = "0x6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"

// each eth_address => get a keystore file => with the following files
const TestKsMortezaFile = "test_morteza--3eb92807f1f91a8d4d85bc908c7f86dcddb1df57"
const TestKsBabaYagaFile = "test_babayaga--6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
// test keystore password
const TestKsAccountsPwd = "security123"
