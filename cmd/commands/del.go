package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/xujiajun/nutsdb"
)

// DelCmd represents the delkv command
var DelCmd = &cobra.Command{
	Use:     "delkv [KEY]",
	Short:   "Remove a stored key.",
	Long:    ``,
	Aliases: []string{"d"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.DB.Update(
			func(tx *nutsdb.Tx) error {
				key := []byte(args[0])
				return tx.Delete(database.Bucket, key)
			}); err != nil {
			fmt.Printf("Error deleting value: %s\n", err.Error())
		}
	},
}
