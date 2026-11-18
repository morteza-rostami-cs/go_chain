package wallet

import "path/filepath"

const keystoreDirName = "keystore"

// keys storage path
func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}