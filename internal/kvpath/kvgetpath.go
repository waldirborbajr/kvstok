// Package kvpath provides utilities for resolving file system paths used by kvstok.
package kvpath

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"

	"github.com/waldirborbajr/kvstok/internal/must"
)

// GetKVPath returns the path of the current executable.
func GetKVPath() string {
	pwd, err := os.Executable()
	must.Must(err, "GetKVPath() - getting current path.")
	return pwd
}

// GetKVHomeDir returns the $HOME directory of the current user.
func GetKVHomeDir() string {
	home, err := os.UserHomeDir()
	must.Must(err, "GetKVHomeDir() - getting $HOME path.")
	return home
}

// GenHash generates a SHA-256 hash of the given file and returns it as a hex string.
func GenHash(filename string) string {
	f, err := os.Open(filename)
	must.Must(err, "GenHash() - generating hash code")
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %s\n", err)
		}
	}()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
