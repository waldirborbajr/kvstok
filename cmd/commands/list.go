package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/xujiajun/nutsdb"
)

// LstCmd represents the lstkv command
var LstCmd = &cobra.Command{
	Use:     "listkv",
	Short:   "List all keys values pairs.",
	Long:    ``,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.DB.View(
			func(tx *nutsdb.Tx) error {
				if nodes, err := tx.GetAll(database.Bucket); err != nil {
					return err
				} else {
					for _, node := range nodes {
						fmt.Println(string(node.Key), " ", string(node.Value))
					}
				}

				return nil
			}); err != nil {
			fmt.Printf("Error listing keys database keys must be empty: %s", err.Error())
		}
	},
}
