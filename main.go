package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	//convert string, to slice of bytes
	//SHA256
	balanceHash := sha256.Sum256([]byte("| 1 | morteza | 99895 |"))
	//52f238321451fe0f4fbe4e94b3c489ba88b9e8970c5f8daaabe3d0308f317583s


	fmt.Printf("%x\n", balanceHash)

	// changing one number, generates complete different hash
	balanceHashDiff := sha256.Sum256([]byte("| 1 | morteza | 99896 |"))

	fmt.Printf("%x\n", balanceHashDiff)
	// ade38b693c1da84dd00e808e1e788a6086d6fe17883f5247801f46eb53467a2e

}