// internal/clipboard/clipboard.go
package clipboard

import (
	"fmt"

	"golang.design/x/clipboard"
)

// Copy copies text to the clipboard
func Copy(text string) error {
	if text == "" {
		return fmt.Errorf("nothing to copy")
	}

	// Initialize clipboard support once
	clipboard.Init()

	err := clipboard.Write(clipboard.FmtText, []byte(text))
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// CopyWithConfirmation copies text and prints a confirmation message
func CopyWithConfirmation(text, key string) error {
	if err := Copy(text); err != nil {
		return err
	}

	fmt.Printf("✅ Key '%s' copied to the clipboard!\n", key)
	return nil
}
