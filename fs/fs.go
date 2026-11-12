package fs

import (
	"os"
	"os/user"
	"path"
	"strings"
)

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