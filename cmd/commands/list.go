package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// LstCmd represents the list command
var LstCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all keys values pairs.",
	Long:    ``,
	Aliases: []string{"listkv", "l"},
	Run: func(cmd *cobra.Command, args []string) {
		store, err := database.GetStore()
		must.Must(err, "LstCmd() - failed to open store")

		entries, err := store.List()
		must.Must(err, "LstCmd() - failed to list entries")

		for k, e := range entries {
			fmt.Println(k, " ", e.Value)
		}
	},
}
