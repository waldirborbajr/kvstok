package commands

import (
	"fmt"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// LstCmd represents the lstkv command
var LstCmd = &cobra.Command{
	Use:     "{l}istkv",
	Short:   "List all keys values pairs.",
	Long:    ``,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		err := database.DB.View(
			func(tx *nutsdb.Tx) error {
				if keys, values, err := tx.GetAll(database.Bucket); err != nil {
					return err
				} else {
					n := len(keys)
					for i := 0; i < n; i++ {
						fmt.Println(string(keys[i]), " ", string(values[i]))
					}
				}

				return nil
			})

		must.Must(err, "LstCmd() - key not found or datababse must be empty.")
	},
}
