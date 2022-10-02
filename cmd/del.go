package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/xujiajun/nutsdb"
)

var deleteCmd = &cobra.Command{
	Use:   "delkv key",
	Short: "Remove a key.",
	Long:  "Remove a key previously stored into database.",
	Args:  cobra.MinimumNArgs(1),
	Run:   deleteVal,
}

func deleteVal(cmd *cobra.Command, args []string) {
	if err := database.DB.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(args[0])
			return tx.Delete(database.Bucket, key)
		}); err != nil {
		fmt.Printf("Error deleting value: %s\n", err.Error())
	}
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
