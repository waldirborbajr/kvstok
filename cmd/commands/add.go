package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// AddCmd represents the addkv command
var AddCmd = &cobra.Command{
	Use:     "{a}ddkv [KEY] [VALUE]",
	Short:   "Add or update a value for a key.",
	Long:    ``,
	Aliases: []string{"a"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("addkv requires two parameters [key] and [value]. Please try again")
		}
		return nil
	},
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: kvstok add <key> <value>")
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

	fmt.Printf("✅ Key '%s' saved successfully!\n", key)
	return nil
}
