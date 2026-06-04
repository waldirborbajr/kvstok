package commands

import (
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// DelCmd represents the del command
var DelCmd = &cobra.Command{
	Use:     "del [KEY]",
	Short:   "Remove a stored key.",
	Long:    ``,
	Aliases: []string{"delkv", "d"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		store, err := database.GetStore()
		must.Must(err, "DelCmd() - failed to open store")

		if err := store.Delete(args[0]); err != nil {
			must.Must(err, "DelCmd() - oops! Huston, we have a problem deleting keys. The key does not exist or database must be empty.")
		}
	},
}
