package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"go_chain/database"
	"os"

	// "go_chain/database"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

// sub dir name, for storing encrypted keys
const keystoreDirName = "keystore"

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

// load privKey from file, and sign tx with it
func SignTxWithKeystoreAccount(
	tx database.Tx, // tx to sign
	acc	database.Account, // eth address of the account that should sign it
	pwd	string, // passphrase, to unlock the keystore
	keystoreDir	string,// key dir path to keystore
) (database.SignedTx, error) {

	// load and config
	ksManager := keystore.NewKeyStore(
		keystoreDir, 
		keystore.StandardScryptN, 
		keystore.StandardScryptP,
	)

	// keystore can hold multiple accounts => find the one we want to sign with
	ksAccount, err := ksManager.Find(accounts.Account{Address: acc})
	if err != nil {
		return database.SignedTx{}, err
	}

	/*
	keystore/
    UTC--...--account1
    UTC--...--account2
	*/

	// read the file into json
	ksAccountJson, err := os.ReadFile(ksAccount.URL.Path)
	if err != nil {
		return database.SignedTx{}, err
	}

	// use password to decrypt the privKey, for signing
	key, err := keystore.DecryptKey(ksAccountJson, pwd)
	if err != nil {
		return database.SignedTx{}, err
	}

	// sign the tx
	signedTx, err := SignTx(tx, key.PrivateKey)
	if err != nil {
		return database.SignedTx{}, err
	}
	
	return signedTx, nil
}

// sign and return SignedTx
func SignTx(
	tx database.Tx,
	privKey *ecdsa.PrivateKey,	
) (database.SignedTx, error) {
	// get tx raw json bytes
	rawTx, err := tx.Encode() 
	if err != nil {
		return database.SignedTx{}, err
	}

	// sign tx
	sig, err := Sign(rawTx, privKey)
	if err != nil {
		return database.SignedTx{}, err
	}

	// return a SignedTx
	return database.NewSignedTx(tx, sig), nil
}

// take: a message []byte, and, private key
// return a [65]byte signature
func Sign(
	msg []byte,
	privKey *ecdsa.PrivateKey,
) ([]byte, error) {

	// hash the message or tx or whatever bytes, we need the hash for signing
	// returns: [32]byte keccak-256 hash
	// msgHash := crypto.Keccak256(msg)
	
	// use the same hashing as tx.Hash()
	var msgHash = sha256.Sum256(msg)

	// sign the message hash
	sig, err := crypto.Sign(msgHash[:], privKey)
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
	// var msgHash = crypto.Keccak256(msg)

	// use sha256 => to match the IsAuthentic()
	var msgHash = sha256.Sum256(msg)
	
	// verify private_key signing, msg + sig = public_key
	publicKey, err := crypto.SigToPub(msgHash[:], sig)

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

// this func is for generating a random miner account for testing
// so: key.address => gives you a real eth address, derived from a private key 
func NewRandomKey() (*keystore.Key, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	id, _ := uuid.NewRandom()
	// this has eth address => key.address
	key := &keystore.Key{
		Id:	id,
		Address:	crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key, nil
}