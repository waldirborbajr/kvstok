// internal/clipboard/clipboard.go
package clipboard

import (
	"fmt"
	"sync"
	"time"

	"golang.design/x/clipboard"
)

var initOnce sync.Once
var initErr error

func initClipboard() error {
	initOnce.Do(func() {
		initErr = clipboard.Init()
	})
	return initErr
}

// Copy copies text to the clipboard
func Copy(text string) error {
	if text == "" {
		return fmt.Errorf("nothing to copy")
	}

	if err := initClipboard(); err != nil {
		return fmt.Errorf("clipboard initialization failed: %w", err)
	}

	done := clipboard.Write(clipboard.FmtText, []byte(text))
	select {
	case <-done:
		// successfully copied to clipboard
	case <-time.After(200 * time.Millisecond):
		go func() { <-done }()
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
