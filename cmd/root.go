package cmd

import (
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/cmd/commands"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/must"
	"github.com/waldirborbajr/kvstok/pkg/config"
	"github.com/xujiajun/nutsdb"
)

// Size of database to store key/value
const DBSIZE = 2048 * 2048

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kvstok",
	Short:   "KVStoK is a CLI-based KEY VALUE storage.",
	Version: config.Verzion,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	must.Must(rootCmd.Execute())
}

func init() {
	// Import config
	initConfig()

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.DisableSuggestions = true

	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.DelCmd)
	rootCmd.AddCommand(commands.GetCmd)
	rootCmd.AddCommand(commands.LstCmd)
}

func initConfig() {
	homePath := kvpath.GetKVHomeDir() + "/" + database.DBName

	database.DB, _ = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(homePath),
		nutsdb.WithSegmentSize(DBSIZE),
	)
}
