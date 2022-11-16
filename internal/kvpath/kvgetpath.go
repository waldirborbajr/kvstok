package kvpath

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"

	"github.com/waldirborbajr/kvstok/internal/must"
)

// Get current path and returns
func GetKVPath() string {
	pwd, err := os.Executable()
	must.Must(err)
	// if err != nil {
	// 	fmt.Printf("Error trying to get current path. %s", err.Error())
	// 	os.Exit(-1)
	// }

	return pwd
}

// Get $HOME path of user and returns
func GetKVHomeDir() string {
	home, err := os.UserHomeDir()
	must.Must(err)
	// if err != nil {
	// 	fmt.Printf("Error acquiring Home Dir path: %s", err.Error())
	// }

	return home
}

// Generate HASH of a given file
func GenHash(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
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
