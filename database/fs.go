package database

import (
	"os"
	"path/filepath"
)

// create a /data folder if it does not exists
func initDataDirIfNotExists(dataDir string) error {
	// check if /data/genesis.json file exists or not
	// if exists:: return nil -- do not create anything
	//getGenesisJsonFilePath() -- get path to genesis.json file
	if fileExist(getGenesisJsonFilePath(dataDir)) {
		return nil
	}

	// create /data directory and any missing parent folders
	if err := os.MkdirAll(getDatabaseDirPath(dataDir), os.ModePerm); err != nil { return err }

	// writes data to genesis.json
	if err := writeGenesisToDisk(getGenesisJsonFilePath(dataDir)); err != nil { return err }

	// create an empty block.db file
	if err := writeEmptyBlocksDbToDisk(getBlocksDbFilePath(dataDir)); err != nil { return err }

	return nil
}

// functions for getting file path to different files
func getDatabaseDirPath(dataDir string) string {
	// return filepath.Join(dataDir, "data")
	return filepath.Join(dataDir)
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDbFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "block.db")
}

// check whether a file exists
func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	// err => just means something went wrong, getting info
	// IsNotExist() -- check if err is: path does not exists
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// check if a dir exists
func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)

	// directory exists
	if err == nil { return true, nil }

	// check if path does not exists
	if os.IsNotExist(err) { return false, nil }

	// directory exists
	return true, err
}

// create an empty block.db file
func writeEmptyBlocksDbToDisk(path string) error{
	return os.WriteFile(
		path,
		[]byte(""), // write zero bytes
		os.ModePerm,// max permitted bits, depends on OS
	)
}

// func writeGenesisToDisk(path string) error {

// 	return nil
// } 



