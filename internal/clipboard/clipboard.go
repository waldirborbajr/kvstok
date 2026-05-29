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

	// Write returns a done channel, not an error — discard it
	<-clipboard.Write(clipboard.FmtText, []byte(text))

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
