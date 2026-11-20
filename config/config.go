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
	var pubKeyBytes = elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	// skip the prefix 0x04 
	var pubKeyBytesHash = crypto.Keccak256(pubKeyBytes[1:])
	// account address => is the last 20 bytes -- take bytes 12..32 (the last 20)
	var account = common.BytesToAddress(pubKeyBytesHash[:12])

	return account
}