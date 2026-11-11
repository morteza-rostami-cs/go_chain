package database

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
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

// handle ~/ this type of path, expanding it to /home/something
// root path, is not effected => eg: ./data
func ExpandPath(address string) string {
	// check for remote locations: user@host:/some/address
	// Index() returns -1 , if does not exists!
	// i > 0 => means char has to show up somewhere after the inx[0]
	if i := strings.Index(address, ":"); i > 0 {
		// found first occurrence of ":"
		return address
	}

	//user@host:/some/address -- we want local address or user@host:/some/address
	if i := strings.Index(address, "@"); i > 0 {
		return address
	}

	// if: strings starts with: ~/ or ~\\
	// ~/projects/blockchain
	// homeDir() :: /home/me
	if strings.HasPrefix(address, "~/") || strings.HasPrefix(address, "~\\") {
		// expand this ~/ to /home/something/...
		if home := homeDir(); home != "" {
			// remove ~/ and get something like: /home/me/tmc
			address = home + address[1:]
		}
	}

	// ExpandEnv() :: replace env variables: eg: $HOME = /home/me
	// Clean() :: clean this /home/me/projects/../.tbb to this:: /home/me/.tbb => It also removes repeated separators and unnecessary . elements.
	return path.Clean(os.ExpandEnv(address))
}

func homeDir() string {
	// read value of HOME= :: which we can set in our command: $HOME/
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// get HOME from current os user
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}

	// fall back
	return ""
}