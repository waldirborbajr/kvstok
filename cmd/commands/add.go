package commands

import (
	"errors"
	"fmt"

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
				return tx.Put(database.Bucket, []byte(args[0]), []byte(args[1]), 0)
			})

		must.Must(err, "AddCmd() - oops! Huston, we have a problem adding/updating keys.")
	},
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("uso: kvstok add <chave> <valor>")
	}

	store, err := GetStore()
	if err != nil {
		return err
	}
	defer store.Close()

	key := args[0]
	value := args[1]

	if err := store.Put(key, value, 0, nil); err != nil {
		return err
	}

	fmt.Printf("✅ Chave '%s' salva com sucesso!\n", key)
	return nil
}
