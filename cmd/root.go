package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/cmd/commands"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/must"
	"github.com/waldirborbajr/kvstok/internal/version"
	"github.com/xujiajun/nutsdb"
)

// Size of database to store key/value
const DBSIZE = 2048 * 2048

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kvstok",
	Short:   "KVStoK is a CLI-based KEY VALUE storage.",
	Version: version.AppVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	must.Must(rootCmd.Execute())
}

// before release 0.2.0 database must be moved to ~/.config/kvstok
func movedb() {
	home := kvpath.GetKVHomeDir() + "/" + database.DBName
	newHome := kvpath.GetKVHomeDir() + "/.config/kvstok/" + database.DBName

	_, err := os.Stat(home)

	if !errors.Is(err, os.ErrNotExist) {
		fmt.Println("Moving database to the new location: ", newHome)
		err := os.Mkdir(kvpath.GetKVHomeDir()+"/.config/kvstok", 0755)
		if err != nil {
			// log.Fatal(err)
		}
		if err := os.Rename(home, newHome); err != nil {
			log.Fatal("Error moving ", err)
		}
	}

}

func init() {

	movedb()

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
}

func initConfig() {
	homePath := kvpath.GetKVHomeDir() + "/.config/kvstok/" + database.DBName

	database.DB, _ = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(homePath),
		nutsdb.WithSegmentSize(DBSIZE),
	)
}
