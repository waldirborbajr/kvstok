package cmd

import (
	"fmt"
	"os"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/cmd/commands"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/must"
	"github.com/waldirborbajr/kvstok/internal/version"
)

// Size of database to store key/value
const DBSIZE = 2048 * 2048

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
	must.Must(rootCmd.Execute(), "Excute() on parsing commands.")
}

// before release 0.2.0 database must be moved to ~/.config/kvstok
// deprecated will be removed on release 0.4.0
func movedb() {
	home := kvpath.GetKVHomeDir() + "/" + database.DBName
	newHome := kvpath.GetKVHomeDir() + "/.config/kvstok/" + database.DBName

	if _, err := os.Stat(home); err == nil {
		fmt.Println("Moving database to the new location: ", newHome)
		if _, err := os.Stat(kvpath.GetKVHomeDir() + "/.config/kvstok"); err != nil {
			os.Mkdir(kvpath.GetKVHomeDir()+"/.config/kvstok", 0600)
		}
		if err := os.Rename(home, newHome); err != nil {
			must.Must(err, "On move database.")
		}
	}
}

func init() {
	// TODO: remove on release 0.4.0
	movedb()
	// /TODO

	// Import config
	initConfig()

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.DisableSuggestions = true

	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.DelCmd)
	rootCmd.AddCommand(commands.GetCmd)
	rootCmd.AddCommand(commands.LstCmd)
	rootCmd.AddCommand(commands.ExpCmd)
	rootCmd.AddCommand(commands.ImpCmd)
	rootCmd.AddCommand(commands.TtlCmd)
}

func initConfig() {
	homePath := kvpath.GetKVHomeDir() + "/.config/kvstok/" + database.DBName

	database.DB, _ = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(homePath),
		nutsdb.WithSegmentSize(DBSIZE),
	)
}
