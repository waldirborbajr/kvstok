package cmd

import (
	"fmt" // ✅ ADICIONADO
	"log"
	"os" // ✅ ADICIONADO

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/cmd/commands"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/version"
)

// ✅ VARIÁVEL DECLARADA
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
	// Import config
	initConfig()

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

func initConfig() {
	store, err := database.NewStore("")
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := database.DB.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket)
	}); err != nil {
		log.Fatal(err.Error())
	}
}

func preRun(cmd *cobra.Command, args []string) error {
	store, err := database.NewStore("")
	if err != nil {
		return err
	}
	defer store.Close()

	// Load the salt if it exists
	_ = store.LoadMasterSalt()

	// If the user provided --master, derive the master key
	if masterPassword != "" {
		if err := store.GetMasterKey().SetMasterPassword(masterPassword); err != nil {
			return fmt.Errorf("invalid master password")
		}
		return nil
	}

	// If the store is not initialized, require init
	if !store.IsMasterPasswordSet() {
		if cmd.Use != "init" {
			fmt.Println("⚠️  kvstok is not initialized yet.")
			fmt.Println("   Run: kvstok init")
			os.Exit(1)
		}
	}

	return nil
}

// GetStore returns a store with the master password salt loaded (used by commands)
func GetStore() (*database.Store, error) {
	store, err := database.NewStore("")
	if err != nil {
		return nil, err
	}

	if err := store.LoadMasterSalt(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}
