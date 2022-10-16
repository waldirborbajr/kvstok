package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/xujiajun/nutsdb"
)

var getCmd = &cobra.Command{
	Use:   "getkv key",
	Short: "Get a value for a key.",
	Args:  cobra.MinimumNArgs(1),
	Run:   getVal,
}

func getVal(cmd *cobra.Command, args []string) {
	if err := database.DB.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(args[0])
			content, err := tx.Get(database.Bucket, key)
			if err != nil {
				fmt.Printf("Error getting value: Key [%s] does not exists \n", string(key))
			}
			fmt.Printf("%s\n", content.Value)
			return nil
		}); err != nil {
	}
}

func init() {
	RootCmd.AddCommand(getCmd)
}
