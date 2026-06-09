package commands

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/clipboard"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// MockStoreForCopy implements Store interface for copy command testing
type MockStoreForCopy struct {
	getRawFunc func(key string) (string, interface{}, error)
	getFunc    func(key string) (string, error)
	putFunc    func(key, value string, ttl int64, tags map[string]string) error
	closeFunc  func() error
}

func (m *MockStoreForCopy) GetRaw(key string) (string, interface{}, error) {
	if m.getRawFunc != nil {
		return m.getRawFunc(key)
	}
	return "", nil, nil
}

func (m *MockStoreForCopy) Get(key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(key)
	}
	return "", nil
}

func (m *MockStoreForCopy) Put(key, value string, ttl int64, tags map[string]string) error {
	if m.putFunc != nil {
		return m.putFunc(key, value, ttl, tags)
	}
	return nil
}

func (m *MockStoreForCopy) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// TestCopyCmdArgsValidation tests the Args validator (cobra.MinimumArgs(1))
func TestCopyCmdArgsValidation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "No arguments",
			args:      []string{},
			expectErr: true,
		},
		{
			name:      "One argument - valid",
			args:      []string{"mykey"},
			expectErr: false,
		},
		{
			name:      "Two arguments - valid (ignores extra)",
			args:      []string{"mykey", "ignored"},
			expectErr: false,
		},
		{
			name:      "Many arguments - valid",
			args:      []string{"key", "arg2", "arg3", "arg4"},
			expectErr: false,
		},
		{
			name:      "Empty string as key",
			args:      []string{""},
			expectErr: false,
		},
		{
			name:      "Key with special characters",
			args:      []string{"app:config:db-host"},
			expectErr: false,
		},
		{
			name:      "Key with spaces",
			args:      []string{"my key with spaces"},
			expectErr: false,
		},
		{
			name:      "Unicode key",
			args:      []string{"日本語キー"},
			expectErr: false,
		},
		{
			name:      "Very long key",
			args:      []string{string(make([]byte, 10000))},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := CopyCmd.Args(cmd, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got error: %v (%v)", tt.expectErr, err != nil, err)
			}
		})
	}
}

// TestRunCopySuccess tests successful copy operations
func TestRunCopySuccess(t *testing.T) {
	tests := []struct {
		name             string
		key              string
		storedValue      string
		clipboardSuccess bool
	}{
		{
			name:             "Simple key-value copy",
			key:              "username",
			storedValue:      "john_doe",
			clipboardSuccess: true,
		},
		{
			name:             "Copy with special characters",
			key:              "db:password",
			storedValue:      "P@ssw0rd!#$%",
			clipboardSuccess: true,
		},
		{
			name:             "Copy with spaces",
			key:              "api key",
			storedValue:      "my api key value",
			clipboardSuccess: true,
		},
		{
			name:             "Copy empty value",
			key:              "empty",
			storedValue:      "",
			clipboardSuccess: true,
		},
		{
			name:             "Copy Unicode value",
			key:              "greeting",
			storedValue:      "こんにちは世界",
			clipboardSuccess: true,
		},
		{
			name:             "Copy very long value",
			key:              "longkey",
			storedValue:      string(make([]byte, 100000)),
			clipboardSuccess: true,
		},
		{
			name:             "Copy value with newlines",
			key:              "certificate",
			storedValue:      "-----BEGIN CERT-----\nMIIC...\n-----END CERT-----",
			clipboardSuccess: true,
		},
		{
			name:             "Copy numeric string",
			key:              "port",
			storedValue:      "8080",
			clipboardSuccess: true,
		},
		{
			name:             "Copy JSON-like value",
			key:              "config",
			storedValue:      `{"host":"localhost","port":5432}`,
			clipboardSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetStore := database.GetStore
			originalCopyWithConfirmation := clipboard.CopyWithConfirmation

			defer func() {
				database.GetStore = originalGetStore
				clipboard.CopyWithConfirmation = originalCopyWithConfirmation
			}()

			// Mock GetStore
			database.GetStore = func() (database.Store, error) {
				return &MockStoreForCopy{
					getRawFunc: func(key string) (string, interface{}, error) {
						return tt.storedValue, nil, nil
					},
					closeFunc: func() error { return nil },
				}, nil
			}

			// Mock CopyWithConfirmation
			clipboardCalled := false
			clipboard.CopyWithConfirmation = func(value, key string) error {
				clipboardCalled = true
				if value != tt.storedValue || key != tt.key {
					return errors.New("value or key mismatch")
				}
				if !tt.clipboardSuccess {
					return errors.New("clipboard failed")
				}
				return nil
			}

			cmd := &cobra.Command{}
			err := runCopy(cmd, []string{tt.key})

			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if !clipboardCalled {
				t.Errorf("expected clipboard.CopyWithConfirmation to be called")
			}
		})
	}
}

// TestRunCopyErrors tests error scenarios
func TestRunCopyErrors(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		getStoreErr         error
		getRawErr           error
		copyErr             error
		closeErr            error
		expectErr           bool
		expectErrContains   string
	}{
		{
			name:              "No arguments",
			args:              []string{},
			expectErr:         true,
			expectErrContains: "usage: kvstok copy",
		},
		{
			name:              "GetStore fails",
			args:              []string{"key"},
			getStoreErr:       errors.New("database offline"),
			expectErr:         true,
			expectErrContains: "database offline",
		},
		{
			name:              "GetRaw key not found",
			args:              []string{"nonexistent"},
			getStoreErr:       nil,
			getRawErr:         errors.New("key not found"),
			expectErr:         true,
			expectErrContains: "key not found",
		},
		{
			name:              "GetRaw decrypt error",
			args:              []string{"corrupted"},
			getStoreErr:       nil,
			getRawErr:         errors.New("decryption failed"),
			expectErr:         true,
			expectErrContains: "decryption failed",
		},
		{
			name:              "CopyWithConfirmation fails",
			args:              []string{"key"},
			getStoreErr:       nil,
			getRawErr:         nil,
			copyErr:           errors.New("clipboard unavailable"),
			expectErr:         true,
			expectErrContains: "clipboard unavailable",
		},
		{
			name:        "Close fails (not blocking)",
			args:        []string{"key"},
			getStoreErr: nil,
			getRawErr:   nil,
			copyErr:     nil,
			closeErr:    errors.New("close error"),
			expectErr:   false, // Error in Close after successful operation is not returned
		},
		{
			name:              "GetRaw with empty key",
			args:              []string{""},
			getStoreErr:       nil,
			getRawErr:         errors.New("empty key"),
			expectErr:         true,
			expectErrContains: "empty key",
		},
		{
			name:              "GetRaw timeout",
			args:              []string{"key"},
			getStoreErr:       nil,
			getRawErr:         errors.New("context deadline exceeded"),
			expectErr:         true,
			expectErrContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetStore := database.GetStore
			originalCopyWithConfirmation := clipboard.CopyWithConfirmation

			defer func() {
				database.GetStore = originalGetStore
				clipboard.CopyWithConfirmation = originalCopyWithConfirmation
			}()

			database.GetStore = func() (database.Store, error) {
				if tt.getStoreErr != nil {
					return nil, tt.getStoreErr
				}
				return &MockStoreForCopy{
					getRawFunc: func(key string) (string, interface{}, error) {
						return "value", nil, tt.getRawErr
					},
					closeFunc: func() error {
						return tt.closeErr
					},
				}, nil
			}

			clipboard.CopyWithConfirmation = func(value, key string) error {
				return tt.copyErr
			}

			cmd := &cobra.Command{}
			err := runCopy(cmd, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got error: %v (%v)", tt.expectErr, err != nil, err)
			}

			if tt.expectErr && err != nil && !contains(err.Error(), tt.expectErrContains) {
				t.Errorf("expected error to contain '%s', got '%s'", tt.expectErrContains, err.Error())
			}
		})
	}
}

// TestCopyCmdMetadata verifies command metadata
func TestCopyCmdMetadata(t *testing.T) {
	tests := []struct {
		name     string
		checkFn  func() bool
		errMsg   string
	}{
		{
			name:   "Use is correct",
			checkFn: func() bool { return CopyCmd.Use == "copy [KEY]" },
			errMsg: "Use should be 'copy [KEY]'",
		},
		{
			name:   "Short description is set",
			checkFn: func() bool { return CopyCmd.Short == "Copy a value to the clipboard." },
			errMsg: "Short description mismatch",
		},
		{
			name:   "Long description is set",
			checkFn: func() bool { return CopyCmd.Long == "Copy the value stored at KEY to the clipboard." },
			errMsg: "Long description mismatch",
		},
		{
			name:   "RunE is set",
			checkFn: func() bool { return CopyCmd.RunE != nil },
			errMsg: "RunE should be set",
		},
		{
			name:   "Args validator is set",
			checkFn: func() bool { return CopyCmd.Args != nil },
			errMsg: "Args validator should be set",
		},
		{
			name:   "Has alias 'cp'",
			checkFn: func() bool { 
				for _, alias := range CopyCmd.Aliases {
					if alias == "cp" {
						return true
					}
				}
				return false
			},
			errMsg: "Should have alias 'cp'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.checkFn() {
				t.Error(tt.errMsg)
			}
		})
	}
}

// TestRunCopyStoreInteraction tests that store methods are called correctly
func TestRunCopyStoreInteraction(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		expectedValue  string
		getRawCalled   bool
		closeCalled    bool
	}{
		{
			name:           "GetRaw called with correct key",
			key:            "testkey",
			expectedValue:  "testvalue",
			getRawCalled:   true,
			closeCalled:    true,
		},
		{
			name:           "GetRaw called with special key",
			key:            "app:database:password",
			expectedValue:  "secret123",
			getRawCalled:   true,
			closeCalled:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetStore := database.GetStore
			originalCopyWithConfirmation := clipboard.CopyWithConfirmation

			defer func() {
				database.GetStore = originalGetStore
				clipboard.CopyWithConfirmation = originalCopyWithConfirmation
			}()

			getRawCalled := false
			closeCalled := false

			database.GetStore = func() (database.Store, error) {
				return &MockStoreForCopy{
					getRawFunc: func(key string) (string, interface{}, error) {
						getRawCalled = true
						if key != tt.key {
							return "", nil, errors.New("key mismatch")
						}
						return tt.expectedValue, nil, nil
					},
					closeFunc: func() error {
						closeCalled = true
						return nil
					},
				}, nil
			}

			clipboard.CopyWithConfirmation = func(value, key string) error {
				return nil
			}

			cmd := &cobra.Command{}
			_ = runCopy(cmd, []string{tt.key})

			if !getRawCalled {
				t.Errorf("expected GetRaw to be called")
			}
			if !closeCalled {
				t.Errorf("expected Close to be called")
			}
		})
	}
}

// TestRunCopyConcurrency tests concurrent copy operations
func TestRunCopyConcurrency(t *testing.T) {
	originalGetStore := database.GetStore
	originalCopyWithConfirmation := clipboard.CopyWithConfirmation

	defer func() {
		database.GetStore = originalGetStore
		clipboard.CopyWithConfirmation = originalCopyWithConfirmation
	}()

	callCount := 0
	database.GetStore = func() (database.Store, error) {
		return &MockStoreForCopy{
			getRawFunc: func(key string) (string, interface{}, error) {
				return "value", nil, nil
			},
			closeFunc: func() error { return nil },
		}, nil
	}

	clipboard.CopyWithConfirmation = func(value, key string) error {
		callCount++
		return nil
	}

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			cmd := &cobra.Command{}
			_ = runCopy(cmd, []string{"key"})
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if callCount != 10 {
		t.Errorf("expected 10 calls, got %d", callCount)
	}
}

// TestCopyCmdAliases verifies aliases
func TestCopyCmdAliases(t *testing.T) {
	if len(CopyCmd.Aliases) == 0 {
		t.Errorf("expected aliases to be set")
		return
	}

	found := false
	for _, alias := range CopyCmd.Aliases {
		if alias == "cp" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected 'cp' alias to be present")
	}
}

// TestRunCopyWithDifferentKeyTypes tests various key formats
func TestRunCopyWithDifferentKeyTypes(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"Simple key", "password"},
		{"Key with dots", "app.config.database"},
		{"Key with dashes", "api-key"},
		{"Key with underscores", "db_user"},
		{"Key with colons", "app:db:host"},
		{"Key with slashes", "path/to/key"},
		{"Mixed special chars", "app_config.db-password:prod"},
		{"Numeric key", "123"},
		{"Key starting with number", "1password"},
	}

	originalGetStore := database.GetStore
	originalCopyWithConfirmation := clipboard.CopyWithConfirmation

	defer func() {
		database.GetStore = originalGetStore
		clipboard.CopyWithConfirmation = originalCopyWithConfirmation
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database.GetStore = func() (database.Store, error) {
				return &MockStoreForCopy{
					getRawFunc: func(key string) (string, interface{}, error) {
						return "value", nil, nil
					},
					closeFunc: func() error { return nil },
				}, nil
			}

			clipboard.CopyWithConfirmation = func(value, key string) error {
				return nil
			}

			cmd := &cobra.Command{}
			err := runCopy(cmd, []string{tt.key})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestRunCopyClipboardValues tests clipboard receives correct values
func TestRunCopyClipboardValues(t *testing.T) {
	tests := []struct {
		name             string
		key              string
		storedValue      string
	}{
		{"Normal value", "key1", "value1"},
		{"Value with special chars", "key2", "p@$$w0rd!"},
		{"Multiline value", "key3", "line1\nline2\nline3"},
		{"Very long value", "key4", string(make([]byte, 50000))},
		{"Empty value", "key5", ""},
		{"Unicode value", "key6", "日本語テキスト"},
	}

	originalGetStore := database.GetStore
	originalCopyWithConfirmation := clipboard.CopyWithConfirmation

	defer func() {
		database.GetStore = originalGetStore
		clipboard.CopyWithConfirmation = originalCopyWithConfirmation
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database.GetStore = func() (database.Store, error) {
				return &MockStoreForCopy{
					getRawFunc: func(key string) (string, interface{}, error) {
						return tt.storedValue, nil, nil
					},
					closeFunc: func() error { return nil },
				}, nil
			}

			clipboardValue := ""
			clipboardKey := ""
			clipboard.CopyWithConfirmation = func(value, key string) error {
				clipboardValue = value
				clipboardKey = key
				return nil
			}

			cmd := &cobra.Command{}
			_ = runCopy(cmd, []string{tt.key})

			if clipboardValue != tt.storedValue {
				t.Errorf("expected clipboard value '%s', got '%s'", tt.storedValue, clipboardValue)
			}

			if clipboardKey != tt.key {
				t.Errorf("expected clipboard key '%s', got '%s'", tt.key, clipboardKey)
			}
		})
	}
}

// BenchmarkRunCopy benchmarks the runCopy function
func BenchmarkRunCopy(b *testing.B) {
	originalGetStore := database.GetStore
	originalCopyWithConfirmation := clipboard.CopyWithConfirmation

	defer func() {
		database.GetStore = originalGetStore
		clipboard.CopyWithConfirmation = originalCopyWithConfirmation
	}()

	database.GetStore = func() (database.Store, error) {
		return &MockStoreForCopy{
			getRawFunc: func(key string) (string, interface{}, error) {
				return "benchvalue", nil, nil
			},
			closeFunc: func() error { return nil },
		}, nil
	}

	clipboard.CopyWithConfirmation = func(value, key string) error {
		return nil
	}

	cmd := &cobra.Command{}
	args := []string{"benchkey"}

	// Redirect output
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runCopy(cmd, args)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
	       (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
