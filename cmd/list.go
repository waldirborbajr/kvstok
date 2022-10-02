package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/xujiajun/nutsdb"
)

var listCmd = &cobra.Command{
	Use:   "listkv [(-o|--output=)json|yaml]",
	Short: "List all keys values pairs.",
	Long:  "List all keys values pairs stored into database, you can export to file too informing [output] option.",
	Run:   listVal,
}

func listVal(cmd *cobra.Command, args []string) {
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
}

func init() {
	RootCmd.AddCommand(listCmd)
}
