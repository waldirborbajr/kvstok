// cmd/commands/tag.go - NOVO ARQUIVO

package commands

import (
	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var TagCmd = &cobra.Command{
	Use:   "tag [add|remove|list] [KEY] [TAG...]",
	Short: "Manage tags for a key.",
	Run: func(cmd *cobra.Command, args []string) {
		subcommand := args[0]
		key := args[1]

		switch subcommand {
		case "add":
			tags := args[2:]
			database.DB.Update(func(tx *nutsdb.Tx) error {
				// Fetch existing entry
				// Add new tags to existing list
				// Update in DB
				return nil
			})
		case "remove":
			// Similar logic
		case "list":
			// Display tags for key
		}
	},
}

// Modificar list.go para filtrar por tags
func filterByTags(keys [][]byte, values [][]byte, tags []string) [][]byte {
	// Filter logic: buscar entries com TODAS as tags solicitadas
	// Return filtered keys
}
