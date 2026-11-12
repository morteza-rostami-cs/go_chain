package database

import (
	"bufio"
	"encoding/json"
	"os"
)

// get a block_hash and returns blocks after it
func GetBlocksAfter(
	blockHash Hash,
	dataDir string,
) ([]Block, error) {
	// open /data/block.db file
	file, err := os.OpenFile(
		getBlocksDbFilePath(dataDir),
		os.O_RDONLY,// read only
		0600,
	)

	if err != nil { return nil, err }
	defer file.Close()

	// store blocks we find here
	var blocks = make([]Block, 0)
	// when we find the target block_hash
	var shouldStartCollecting = false

	// if blockHash [0 0 0 0 ...] empty, target is genesisBlock, so start grabbing the rest of blocks
	// basically: i have zero block, give me all you have. from genesis_block
	if blockHash.IsEmpty() { shouldStartCollecting = true }

	// scanner for going over blocks
	var scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil { return nil, err }

		// store json_block in blockFS
		var blockFs BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blockFs)
		if err != nil { return nil, err }

		if shouldStartCollecting {
			// push the rest of blocks in
			blocks = append(blocks, blockFs.Value)
		}

		// if blockFs.key == hash => we found the block
		if blockFs.Key == blockHash {
			shouldStartCollecting = true
		}
	}

	return blocks, nil
}