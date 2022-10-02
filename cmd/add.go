package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/xujiajun/nutsdb"
)

var addCmd = &cobra.Command{
	Use:   "addkv [KEY] [VALUE]",
	Short: "Add or Update a value for a key.",
	Long:  "Add or Update a value for a key, be careful using this to avoid lose any information stored",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("addkv requires two parameters [key] and [value]. Please try it again")
		}
		return nil
	},
	Run: addVal,
}

func addVal(cmd *cobra.Command, args []string) {
	if err := database.DB.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(args[0])
			val := []byte(args[1])
			return tx.Put(database.Bucket, key, val, 0)
		}); err != nil {
		fmt.Printf("Error saving value: %s\n", err.Error())
	}
}

func init() {
	RootCmd.AddCommand(addCmd)
}
