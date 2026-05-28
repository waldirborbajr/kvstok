// cmd/commands/init.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kvstok by configuring the master password",
	Long: `Initialize kvstok by creating the master password required to encrypt all stored data.

This password will protect all your secrets. Keep it safe!`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	store, err := database.NewStore("")
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()

	if store.IsMasterPasswordSet() {
		fmt.Println("⚠️  kvstok is already initialized.")
		fmt.Println("   Use 'kvstok master change' to change the master password once supported.")
		return nil
	}

	fmt.Println("🔐 kvstok Master Password Setup")
	fmt.Println("=========================================")
	fmt.Println("This password will protect all of your secrets.")
	fmt.Println("")

	password, err := readPassword("Enter the master password: ")
	if err != nil {
		return err
	}

	if len(password) < 8 {
		return fmt.Errorf("the master password must be at least 8 characters")
	}

	confirm, err := readPassword("Confirm the master password: ")
	if err != nil {
		return err
	}

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	if err := store.SetMasterPassword(password); err != nil {
		return fmt.Errorf("failed to set master password: %w", err)
	}

	fmt.Println("\n✅ kvstok initialized successfully!")
	fmt.Println("   All stored data will now be encrypted.")
	fmt.Println("")
	fmt.Println("Tip: use the --master flag to avoid typing the password each time:")
	fmt.Println("   kvstok --master YOURPASSWORD add ...")

	return nil
}

// readPassword reads a password securely without echoing input
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Tenta usar terminal sem eco (melhor UX)
	if term.IsTerminal(int(syscall.Stdin)) {
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // nova linha
		return string(bytePassword), err
	}

	// Fallback para ambiente sem terminal (ex: scripts)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	return strings.TrimSpace(password), err
}

// Helper para checar se master password já está configurada
func (s *Store) IsMasterPasswordSet() bool {
	return s.sec.IsMasterPasswordSet()
}
