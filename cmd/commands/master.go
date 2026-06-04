// cmd/commands/master.go
package commands

import (
	"errors"
	"fmt"

	"github.com/waldirborbajr/kvstok/internal/database"

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
		store, err := database.GetStore()
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
	Short: "Change the master password",
	RunE:  runMasterChange,
}

func runMasterChange(cmd *cobra.Command, args []string) error {
	store, err := database.GetStore()
	if err != nil {
		return err
	}
	defer store.Close()

	if !store.IsMasterPasswordSet() {
		return errors.New("master password is not configured. Run: kvstok init")
	}

	masterFlag, err := cmd.Root().PersistentFlags().GetString("master")
	if err != nil {
		return err
	}

	if masterFlag == "" {
		currentPassword, err := readPassword("Enter current master password: ")
		if err != nil {
			return err
		}

		if err := store.SetMasterPassword(currentPassword); err != nil {
			return fmt.Errorf("invalid current master password: %w", err)
		}
	}

	newPassword, err := readPassword("Enter new master password: ")
	if err != nil {
		return err
	}

	if len(newPassword) < 8 {
		return fmt.Errorf("the master password must be at least 8 characters")
	}

	confirmPassword, err := readPassword("Confirm new master password: ")
	if err != nil {
		return err
	}

	if newPassword != confirmPassword {
		return errors.New("passwords do not match")
	}

	if err := store.ChangeMasterPassword("", newPassword); err != nil {
		return err
	}

	fmt.Println("✅ Master password changed successfully!")
	fmt.Println("   Your existing secrets have been re-encrypted with the new password.")
	return nil
}

func init() {
	MasterCmd.AddCommand(MasterStatusCmd)
	MasterCmd.AddCommand(MasterChangeCmd)
}
