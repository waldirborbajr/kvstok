package commands

import (
	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// DelCmd represents the delkv command
var DelCmd = &cobra.Command{
	Use:     "{d}elkv [KEY]",
	Short:   "Remove a stored key.",
	Long:    ``,
	Aliases: []string{"d"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := database.DB.Update(
			func(tx *nutsdb.Tx) error {
				key := []byte(args[0])
				return tx.Delete(database.Bucket, key)
			})

		must.Must(err, "DelCmd() - oops! Huston, we have a problem deleting keys. The key does not exist or dataase must be empty.")
	},
}
