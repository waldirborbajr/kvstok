package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
	"github.com/xujiajun/nutsdb"
)

// GetCmd represents the getkv command
var GetCmd = &cobra.Command{
	Use:     "{g}etkv [KEY]",
	Short:   "Get a value for a key.",
	Long:    ``,
	Aliases: []string{"g"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// nolint:staticcheck
		if err := database.DB.Update(
			func(tx *nutsdb.Tx) error {
				key := []byte(args[0])
				content, err := tx.Get(database.Bucket, key)
				must.Must(err)
				// if err != nil {
				// fmt.Printf("Error getting value: Key [%s] does not exists \n", string(key))
				// }
				fmt.Printf("%s\n", content.Value)
				return nil
			}); err != nil {
		}
	},
}
