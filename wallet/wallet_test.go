package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"go_chain/config"
	"go_chain/database"
	"go_chain/fs"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/crypto"
)

// fake passphrase for test
const testKeystoreAccountsPwd = "security123"

// create a keystore and sign a transaction
func TestSignTxWithKeystoreAccount(t *testing.T) {
	// get temp dir
	tmpDir, err := ioutil.TempDir("", "wallet_test")
	if err != nil {
		t.Fatal(err)
	}

	defer fs.RemoveDir(tmpDir)

	// make new keystore accounts
	morteza, err := NewKeyStoreAccount(tmpDir, testKeystoreAccountsPwd)

	if err != nil {
		t.Fatal(err)
	}

	baba, err := NewKeyStoreAccount(tmpDir, testKeystoreAccountsPwd)

	if err != nil {
		t.Fatal(err)
	}

	tx := database.NewTx(morteza, baba, 100, "")
	fmt.Println(tx.From)
	signedTx, err := SignTxWithKeystoreAccount(
		tx, 
		morteza, // signer
		testKeystoreAccountsPwd, 
		GetKeystoreDirPath(tmpDir),
	)
	if err != nil {
		t.Fatal(err)
	}

	// print signed tx
	spew.Dump(signedTx.Encode())

	// check for valid tx => msg + sig = eth_address that is equal to tx.From
	ok, err := signedTx.IsAuthentic()
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("this tx was not signed by 'from' account! something is fishy!??")
	}
}

// now lets check a tx signed by a hacker, not the real tx.From address
func TestSignForgedTxWithKeystoreAccount(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "wallet_test")
	if err != nil { t.Fatal(err) }

	defer fs.RemoveDir(tmpDir)

	// hacker account
	hacker, err := NewKeyStoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil { t.Fatal(err) }

	baba, err := NewKeyStoreAccount(tmpDir, testKeystoreAccountsPwd)
	if err != nil { t.Fatal(err) }

	// tx is From = Babayaga => but is signed by the hacker
	var forgedTx = database.NewTx(baba, hacker, 100, "")
	signedTx, err := SignTxWithKeystoreAccount(
		forgedTx,
		hacker, // signer
		testKeystoreAccountsPwd,// password
		GetKeystoreDirPath(tmpDir),
	)

	if err != nil { t.Fatal(err) }

	ok, err := signedTx.IsAuthentic()
	if err != nil { t.Fatal(err) }

	// it should not be ok!!
	if ok {
		t.Fatal("tx.From was forged! and it should not authenticate!")
	}
}

func TestSignCryptoParams(t *testing.T) {
	// get a private key for test
	privKey, err := ecdsa.GenerateKey(
		crypto.S256(),
		rand.Reader,
	)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(privKey)// print priv_key

	var msg = []byte("the Web3Coach students are awesome")

	sig, err := Sign(msg, privKey)

	if err != nil { t.Fatal(err) }

	if len(sig) != crypto.SignatureLength {
		t.Fatalf("wrong sig length: %d", len(sig))
	}

	// extract 3 components of private_key
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])
	v := new(big.Int).SetBytes([]byte{sig[64]})

	spew.Dump(r, s, v)
}

// test getting public key and account from msg + sig
func TestSign(t *testing.T) {
	// generate a private key
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil { t.Fatal(err) }

	// get public key and account from privKey
	var pubKey = privKey.PublicKey

	var account = config.GetAccountFromPubKey(pubKey)

	// message to sign
	var msg = []byte("the Web3Coach students are awesome")

	// sign with private key
	sig, err := Sign(msg, privKey)

	if err != nil { t.Fatal(err) }

	// verify message + sig = public key
	recoveredPubKey, err := Verify(msg, sig)
	if err != nil { t.Fatal(err) }

	var recoveredAcc = config.GetAccountFromPubKey(*recoveredPubKey)

	// check if the account address recovered by Verify() is the same as original
	if account.Hex() != recoveredAcc.Hex() {
		t.Fatalf(
			"msg was signed by acc %s and sig recovery produced and account %s",
			account.Hex(),
			recoveredAcc.Hex(),
		)
	}

	t.Log(account.Hex())
}

