package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:     "refresh key",
	Short:   "Refresh key stored database",
	Aliases: []string{"r"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		showRefresh(key)
	},
}

func init() {
	// refreshCmd.Flags().StringVarP()
	RootCmd.AddCommand(refreshCmd)
}

func showRefresh(key string) {
	fmt.Println("TODO: implement a refresh option " + key)
}
