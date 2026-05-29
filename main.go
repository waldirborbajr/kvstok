package main

import (
	"github.com/waldirborbajr/kvstok/cmd"
)

func main() {
	// Check if RSA keys exist
	// home := kvpath.GetKVHomeDir()

	// database := home + "/.config/kvstok/.6B7673"
	// hasDatabase := true
	// //
	// if _, err := os.Stat(database); err != nil {
	// 	hasDatabase = false
	// }

	// // If keys are missing, inform user to run init
	// if !hasDatabase {
	// 	fmt.Println("⚠️  KVStoK is not initialized.")
	// 	fmt.Println("   Please execute: kvstok init")
	// 	os.Exit(1)
	// }

	cmd.Execute()
}
