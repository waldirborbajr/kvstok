package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/clipboard"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var copyFlag bool

// GetCmd represents the get command
var GetCmd = &cobra.Command{
	Use:     "get [KEY]",
	Short:   "Get a value for a key.",
	Long:    ``,
	Aliases: []string{"getkv", "g"},
	Args:    cobra.MinimumNArgs(1),
	RunE:    runGet,
}

func init() {
	GetCmd.Flags().BoolVarP(&copyFlag, "copy", "c", false, "Copy value to the clipboard")
}

func runGet(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kvstok get <key> [--copy]")
	}

	key := args[0]

	store, err := database.GetStore()
	if err != nil {
		return err
	}
	defer store.Close()

	value, _, err := store.GetRaw(key)
	if err != nil {
		return err
	}

	if copyFlag {
		return clipboard.CopyWithConfirmation(value, key)
	}

	fmt.Println(value)
	return nil
}
