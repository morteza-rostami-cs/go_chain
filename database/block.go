package database

import (
	// "crypto/sha256"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	// "encoding/json"
	// "fmt"
)

// sha-256 always produces a 256-bit hash
// we can compare Hash == Hash
type Hash [32]byte
//methods

// encode our array of bytes into a slice of hex
// Hash and not *Hash -- cause this is read only, so we pass a copy
//🤪🤪 MarshalText() and UnmarshalText() are Special method:: encoding/json discovers and calls this when encodes or decodes a Hash field. eg: json.Marshal(hash)
// there is this TextMarshaler interface{ MarshalText() } 
//encoding.TextMarshaler interface (i.e., has MarshalText() ([]byte, error))
func (self Hash) MarshalText() ([]byte, error) {
	// hex string
	// self[:] -> creates an slice
	var hx = hex.EncodeToString(self[:])

	return []byte(hx), nil
}

/*
for input "aabbcc...", hex.Decode converts every two hex characters into one byte:

go
"aabbcc" -> []byte{0xaa, 0xbb, 0xcc}

#==================
any Hash field inside a block (header.parent, key) will be serialized as a 64-char hex string in JSON.
*/

// convert hex_bytes (string readable) to raw 32 bytes array, and write it into Hash instance 
func (self *Hash) UnmarshalText(data []byte) error {
	// data: hex/json bytes to decode
	// returns: num of bytes decoded, error
	_, err := hex.Decode(
		self[:], // destination: write decoded bytes here 
		data, // hex/json bytes we want to decode
	)
	// if no error returns nil
	return err
}

/*
# in our json => BlockFS

{
  "hash": "a1b2c3...",
  "block": {
    "header": {
      "parent": "000000...",
      "time": 123456789
    },
    "payload": []
  }
}
*/ 

// Hash(Block 0) -- we use the whole content of a block to create it's hash.
type BlockHeader struct{
	// previous block's hash
	Parent Hash `json:"parent"`
	//Unix timestamps, block created at
	Time	uint64 `json:"time"`
}

type Block struct{
	// block header
	// json tags are defining the name of our fields, inside json file
	Header BlockHeader `json:"header"`
	// batch of transactions in a block
	TXs []Tx	`json:"payload"`
}
// methods

// create a hash [32]byte of the instance of our block
func (self Block) Hash() (Hash, error){
	// convert block_struct to json
	// json is encoded as bytes.
	var blockJson []byte // json bytes
	var err error
	// so: when json.Marshal() sees something like this: 
	// Parent Hash `json:"parent"`
	// it calls Hash.MarshalText -- to convert the Hash bytes to hex
	blockJson, err = json.Marshal(self)

	if err != nil {return Hash{}, err}

	// hash the json, sha256 -- needs json bytes, not go struct!
	// returns a hash, with: 64 hexadecimal characters
	return sha256.Sum256(blockJson), nil
}

// block file system, each block stored in json, block.db row
type BlockFS struct{
	// hash identifying a block
	// the last BlockFS.Key -- holds the hash for entire chain.
	Key Hash `json:"hash"`
	// the block itself
	Value Block `json:"block"`
}

// block constructor function
func NewBlock(parent Hash, time uint64, txs []Tx) Block {
	return Block{
		BlockHeader{parent, time},
		txs,
	}
}