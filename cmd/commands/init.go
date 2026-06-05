// Package commands implements the CLI commands for kvstok.
package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/security"
	"golang.org/x/term"
)

// InitCmd initializes kvstok by configuring the master password.
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kvstok by configuring the master password",
	Long: `Initialize kvstok by creating the master password required to encrypt all stored data.

This password will protect all your secrets. Keep it safe!`,
	RunE: runInit,
}

func runInit(_ *cobra.Command, _ []string) error {
	// Generate RSA keys if they don't exist
	if err := ensureRSAKeys(); err != nil {
		return err
	}

	store, err := database.Init("")
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = database.Close() }()

	// Create bucket for storing secrets
	if err := store.DB().Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, database.Bucket)
	}); err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exist") {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	if store.IsMasterPasswordSet() {
		fmt.Println("⚠️  kvstok is already initialized.")
		fmt.Println("   Use 'kvstok master change' to change the master password.")
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

// ensureRSAKeys generates RSA keys if they don't exist
func ensureRSAKeys() error {
	home := kvpath.GetKVHomeDir()
	configDir := filepath.Join(home, ".config", "kvstok")
	pub := filepath.Join(configDir, "kvstok.pub")
	priv := filepath.Join(configDir, "kvstok.priv")

	// Check if keys already exist
	if _, errPub := os.Stat(pub); errPub == nil {
		if _, errPriv := os.Stat(priv); errPriv == nil {
			// Both keys exist
			return nil
		}
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate RSA keys
	fmt.Println("Generating RSA priv/pub keys pairing...")
	privateKey, publicKey := security.RSAGenerateKey(4096)

	// Write public key
	if err := os.WriteFile(pub, security.PublicKeyToBytes(publicKey), 0600); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	// Write private key
	if err := os.WriteFile(priv, security.PrivateKeyToBytes(privateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	fmt.Println("✅ RSA keys generated successfully")
	return nil
}

// readPassword reads a password securely without echoing input
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Try using a no-echo terminal (better UX)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		bytePassword, err := term.ReadPassword(fd)
		fmt.Println() // newline
		return string(bytePassword), err
	}

	// Fallback para ambiente sem terminal (ex: scripts)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	return strings.TrimSpace(password), err
}
