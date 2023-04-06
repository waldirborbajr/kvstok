package commands

import (
	"errors"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// AddCmd represents the addkv command
var AddCmd = &cobra.Command{
	Use:     "{a}ddkv [KEY] [VALUE]",
	Short:   "Add or Update a value for a key.",
	Long:    ``,
	Aliases: []string{"a"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("addkv requires two parameters [key] and [value]. Please try it again")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := database.DB.Update(
			func(tx *nutsdb.Tx) error {
				key := []byte(args[0])
				val := []byte(args[1])
				return tx.Put(database.Bucket, key, val, 0)
			})

		must.Must(err, "DelCmd() - oops! Huston, we have a problem adding/updating keys.")

		// if err := database.DB.Update(
		// 	func(tx *nutsdb.Tx) error {
		// 		key := []byte(args[0])
		// 		val := []byte(args[1])
		// 		return tx.Put(database.Bucket, key, val, 0)
		// 	}); err != nil {
		// 	fmt.Printf("Error saving value: %s\n", err.Error())
		// }
	},
}
