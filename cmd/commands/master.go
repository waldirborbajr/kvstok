// cmd/commands/master.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var MasterCmd = &cobra.Command{
	Use:   "master",
	Short: "Manage the master password",
}

var MasterStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the master password status",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := GetStore()
		if err != nil {
			return err
		}
		defer store.Close()

		if store.IsMasterPasswordSet() {
			fmt.Println("✅ Master password is configured and active")
		} else {
			fmt.Println("❌ Master password is not configured. Run: kvstok init")
		}
		return nil
	},
}

var MasterChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "Change the master password (coming soon)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔄 Master password change is under development...")
		// Future: re-encrypt all stored data with the new password
	},
}

func init() {
	MasterCmd.AddCommand(MasterStatusCmd)
	MasterCmd.AddCommand(MasterChangeCmd)
}
