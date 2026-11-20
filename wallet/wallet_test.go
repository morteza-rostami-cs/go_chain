package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/crypto"
)

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

	var account = GetAccountFromPubKey(pubKey)

	// message to sign
	var msg = []byte("the Web3Coach students are awesome")

	// sign with private key
	sig, err := Sign(msg, privKey)

	if err != nil { t.Fatal(err) }

	// verify message + sig = public key
	recoveredPubKey, err := Verify(msg, sig)
	if err != nil { t.Fatal(err) }

	var recoveredAcc = GetAccountFromPubKey(*recoveredPubKey)

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

