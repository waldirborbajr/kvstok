package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/cmd/commands"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/version"
)

var masterPassword string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kvstok",
	Short:   "KVStoK is a CLI-based KEY VALUE storage.",
	Long:    ``,
	Version: version.AppVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&masterPassword, "master", "m", "", "Master password for kvstok")
	rootCmd.PersistentPreRunE = preRun

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.DisableSuggestions = true

	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.DelCmd)
	rootCmd.AddCommand(commands.GetCmd)
	rootCmd.AddCommand(commands.LstCmd)
	rootCmd.AddCommand(commands.ExpCmd)
	rootCmd.AddCommand(commands.ImpCmd)
	rootCmd.AddCommand(commands.TtlCmd)
	rootCmd.AddCommand(commands.SearchCmd)
	rootCmd.AddCommand(commands.EnvCmd)
	rootCmd.AddCommand(commands.TagCmd)
	rootCmd.AddCommand(commands.MasterCmd)
	rootCmd.AddCommand(commands.InitCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	// Skip validation for init command
	if cmd.Use == "init" {
		return nil
	}

	if !isInitialized() {
		printInitializationMessage()
		os.Exit(1)
	}

	store, err := database.NewStore("")
	if err != nil {
		return err
	}
	defer store.Close()

	// Load the salt if it exists
	_ = store.LoadMasterSalt()

	// Allow master password from CLI or environment variable
	if masterPassword == "" {
		masterPassword = os.Getenv("KVSTOK_MASTER_PASSWORD")
	}

	if masterPassword != "" {
		if err := store.SetMasterPassword(masterPassword); err != nil {
			return fmt.Errorf("invalid master password: %w", err)
		}
		return nil
	}

	// Initialize bucket for other commands
	if err := store.DB().Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket)
	}); err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exist") {
		return err
	}

	return nil
}

func printInitializationMessage() {
	fmt.Println("⚠️  KVStoK is not initialized.")
	fmt.Println("   Please execute: kvstok init")
}

func isInitialized() bool {
	home := kvpath.GetKVHomeDir()
	dbPath := filepath.Join(home, ".config", "kvstok", database.DBName)
	pubPath := filepath.Join(home, ".config", "kvstok", "kvstok.pub")
	privPath := filepath.Join(home, ".config", "kvstok", "kvstok.priv")
	saltPath := filepath.Join(dbPath, "master.salt")

	if _, err := os.Stat(dbPath); err != nil {
		return false
	}
	if _, err := os.Stat(saltPath); err != nil {
		return false
	}
	if _, err := os.Stat(pubPath); err != nil {
		return false
	}
	if _, err := os.Stat(privPath); err != nil {
		return false
	}

	return true
}

// GetStore returns a store with the master password salt loaded (used by commands)
