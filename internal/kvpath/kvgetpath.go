package kvpath

import (
	"fmt"
	"os"
)

// Get current path and returns
func GetKVPath() string {
	pwd, err := os.Executable()
	if err != nil {
		fmt.Printf("Error trying to get current path. %s", err.Error())
		os.Exit(-1)
	}

	return pwd
}

// Get $HOME path of user and returns
func GetKVHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error acquiring Home Dir path: %s", err.Error())
	}

	return home
}
