package main

import (
	"fmt"
	"os"

	"github.com/waldirborbajr/kvstok/cmd"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
)

func main() {
	// Check if RSA keys exist
	home := kvpath.GetKVHomeDir()
	pub := home + "/.config/kvstok/kvstok.pub"
	priv := home + "/.config/kvstok/kvstok.priv"

	hasPub := true
	hasPriv := true

	if _, err := os.Stat(pub); err != nil {
		hasPub = false
	}

	if _, err := os.Stat(priv); err != nil {
		hasPriv = false
	}

	// If keys are missing, inform user to run init
	if !hasPub || !hasPriv {
		fmt.Println("⚠️  KVStoK is not initialized.")
		fmt.Println("   Please execute: kvstok init")
		os.Exit(1)
	}

	cmd.Execute()
}
