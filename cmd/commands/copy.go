package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/clipboard"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// CopyCmd represents the copy command
var CopyCmd = &cobra.Command{
	Use:     "copy [KEY]",
	Short:   "Copy a value to the clipboard.",
	Long:    `Copy the value stored at KEY to the clipboard.`,
	Aliases: []string{"cp"},
	Args:    cobra.MinimumNArgs(1),
	RunE:    runCopy,
}

func runCopy(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kvstok copy <key>")
	}

	key := args[0]

	store, err := database.GetStore()
	if err != nil {
		return err
	}
	defer store.Close()

	// Get the value from the key
	value, _, err := store.GetRaw(key)
	if err != nil {
		return err
	}

	// Copy the value to the clipboard
	return clipboard.CopyWithConfirmation(value, key)
}
