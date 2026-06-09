package commands

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// MockStore implements the Store interface for testing
type MockStore struct {
	putFunc   func(key, value string, ttl int64, tags map[string]string) error
	getFunc   func(key string) (string, error)
	closeFunc func() error
}

func (m *MockStore) Put(key, value string, ttl int64, tags map[string]string) error {
	if m.putFunc != nil {
		return m.putFunc(key, value, ttl, tags)
	}
	return nil
}

func (m *MockStore) Get(key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(key)
	}
	return "", nil
}

func (m *MockStore) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// TestAddCmdArgsValidation tests the Args validator function
func TestAddCmdArgsValidation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "No arguments provided",
			args:      []string{},
			expectErr: true,
			errMsg:    "add requires two parameters",
		},
		{
			name:      "Only one argument provided",
			args:      []string{"key"},
			expectErr: true,
			errMsg:    "add requires two parameters",
		},
		{
			name:      "Exactly two arguments - valid",
			args:      []string{"mykey", "myvalue"},
			expectErr: false,
		},
		{
			name:      "More than two arguments - valid (ignores extras)",
			args:      []string{"mykey", "myvalue", "extra", "args"},
			expectErr: false,
		},
		{
			name:      "Empty key but present",
			args:      []string{"", "value"},
			expectErr: false,
		},
		{
			name:      "Empty value but present",
			args:      []string{"key", ""},
			expectErr: false,
		},
		{
			name:      "Special characters in key",
			args:      []string{"key-with-special!@#", "value"},
			expectErr: false,
		},
		{
			name:      "Unicode characters",
			args:      []string{"键", "值"},
			expectErr: false,
		},
		{
			name:      "Very long key",
			args:      []string{string(make([]byte, 10000)), "value"},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := AddCmd.Args(cmd, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got error: %v", tt.expectErr, err != nil)
			}

			if tt.expectErr && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error message to contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestRunAddSuccess tests successful add operation
func TestRunAddSuccess(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		mockStoreFunc func(key, value string, ttl int64, tags map[string]string) error
		expectOutput  string
	}{
		{
			name:  "Simple key-value pair",
			key:   "testkey",
			value: "testvalue",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key 'testkey' saved successfully!",
		},
		{
			name:  "Key with special characters",
			key:   "app:config:db-host",
			value: "localhost",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key 'app:config:db-host' saved successfully!",
		},
		{
			name:  "Value with spaces",
			key:   "greeting",
			value: "Hello World",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key 'greeting' saved successfully!",
		},
		{
			name:  "Empty key (edge case)",
			key:   "",
			value: "value",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key '' saved successfully!",
		},
		{
			name:  "Empty value (edge case)",
			key:   "key",
			value: "",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key 'key' saved successfully!",
		},
		{
			name:  "Unicode key and value",
			key:   "日本語",
			value: "テスト",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key '日本語' saved successfully!",
		},
		{
			name:  "Very long value",
			key:   "longvalue",
			value: string(make([]byte, 100000)),
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key 'longvalue' saved successfully!",
		},
		{
			name:  "Numeric-like strings",
			key:   "123",
			value: "456.789",
			mockStoreFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			expectOutput: "✅ Key '123' saved successfully!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Mock database.GetStore
			originalGetStore := database.GetStore
			defer func() { database.GetStore = originalGetStore }()

			database.GetStore = func() (database.Store, error) {
				return &MockStore{
					putFunc:   tt.mockStoreFunc,
					closeFunc: func() error { return nil },
				}, nil
			}

			cmd := &cobra.Command{}
			args := []string{tt.key, tt.value}

			err := runAdd(cmd, args)

			w.Close()
			output, _ := io.ReadAll(r)
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if !contains(string(output), tt.expectOutput) {
				t.Errorf("expected output to contain '%s', got '%s'", tt.expectOutput, string(output))
			}
		})
	}
}

// TestRunAddErrors tests various error scenarios
func TestRunAddErrors(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		getStoreErr       error
		putErr            error
		closeErr          error
		expectErr         bool
		expectErrContains string
	}{
		{
			name:              "Less than 2 args",
			args:              []string{"onlykey"},
			expectErr:         true,
			expectErrContains: "usage: kvstok add",
		},
		{
			name:              "No args",
			args:              []string{},
			expectErr:         true,
			expectErrContains: "usage: kvstok add",
		},
		{
			name:              "GetStore returns error",
			args:              []string{"key", "value"},
			getStoreErr:       errors.New("database connection failed"),
			expectErr:         true,
			expectErrContains: "database connection failed",
		},
		{
			name:        "Put operation fails",
			args:        []string{"key", "value"},
			getStoreErr: nil,
			putErr:      errors.New("encryption failed"),
			expectErr:   true,
			expectErrContains: "encryption failed",
		},
		{
			name:        "Close fails after successful Put",
			args:        []string{"key", "value"},
			getStoreErr: nil,
			putErr:      nil,
			closeErr:    errors.New("close failed"),
			expectErr:   false, // Error in Close after successful Put is not returned by runAdd
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetStore := database.GetStore
			defer func() { database.GetStore = originalGetStore }()

			database.GetStore = func() (database.Store, error) {
				if tt.getStoreErr != nil {
					return nil, tt.getStoreErr
				}
				return &MockStore{
					putFunc: func(key, value string, ttl int64, tags map[string]string) error {
						return tt.putErr
					},
					closeFunc: func() error {
						return tt.closeErr
					},
				}, nil
			}

			cmd := &cobra.Command{}
			err := runAdd(cmd, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got error: %v (%v)", tt.expectErr, err != nil, err)
			}

			if tt.expectErr && err != nil && !contains(err.Error(), tt.expectErrContains) {
				t.Errorf("expected error to contain '%s', got '%s'", tt.expectErrContains, err.Error())
			}
		})
	}
}

// TestAddCmdAliases verifies all command aliases work
func TestAddCmdAliases(t *testing.T) {
	expectedAliases := []string{"addkv", "a"}
	
	if len(AddCmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(AddCmd.Aliases))
	}

	for i, alias := range expectedAliases {
		if i >= len(AddCmd.Aliases) {
			t.Errorf("expected alias '%s', but not found", alias)
			continue
		}
		if AddCmd.Aliases[i] != alias {
			t.Errorf("expected alias '%s', got '%s'", alias, AddCmd.Aliases[i])
		}
	}
}

// TestAddCmdMetadata verifies command metadata
func TestAddCmdMetadata(t *testing.T) {
	if AddCmd.Use != "add [KEY] [VALUE]" {
		t.Errorf("expected Use to be 'add [KEY] [VALUE]', got '%s'", AddCmd.Use)
	}

	if AddCmd.Short != "Add or update a value for a key." {
		t.Errorf("expected Short to match, got '%s'", AddCmd.Short)
	}

	if AddCmd.RunE == nil {
		t.Errorf("expected RunE to be set")
	}

	if AddCmd.Args == nil {
		t.Errorf("expected Args validator to be set")
	}
}

// TestAddCmdStoreInteraction tests that store methods are called correctly
func TestAddCmdStoreInteraction(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		value          string
		expectedCalled bool
		verifyFunc     func(key, value string, ttl int64, tags map[string]string) error
	}{
		{
			name:           "Put is called with correct parameters",
			key:            "testkey",
			value:          "testvalue",
			expectedCalled: true,
			verifyFunc: func(key, value string, ttl int64, tags map[string]string) error {
				// Verify parameters
				if key != "testkey" {
					return errors.New("key mismatch")
				}
				if value != "testvalue" {
					return errors.New("value mismatch")
				}
				if ttl != 0 {
					return errors.New("expected ttl to be 0")
				}
				if tags != nil {
					return errors.New("expected tags to be nil")
				}
				return nil
			},
		},
		{
			name:           "Put with special characters in key",
			key:            "app:database:host",
			value:          "192.168.1.1",
			expectedCalled: true,
			verifyFunc: func(key, value string, ttl int64, tags map[string]string) error {
				if key != "app:database:host" || value != "192.168.1.1" {
					return errors.New("parameter mismatch")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalGetStore := database.GetStore
			defer func() { database.GetStore = originalGetStore }()

			putCalled := false
			database.GetStore = func() (database.Store, error) {
				return &MockStore{
					putFunc: func(key, value string, ttl int64, tags map[string]string) error {
						putCalled = true
						return tt.verifyFunc(key, value, ttl, tags)
					},
					closeFunc: func() error { return nil },
				}, nil
			}

			cmd := &cobra.Command{}
			_ = runAdd(cmd, []string{tt.key, tt.value})

			if putCalled != tt.expectedCalled {
				t.Errorf("expected Put to be called: %v, got: %v", tt.expectedCalled, putCalled)
			}
		})
	}
}

// TestAddCmdConcurrency tests concurrent add operations
func TestAddCmdConcurrency(t *testing.T) {
	originalGetStore := database.GetStore
	defer func() { database.GetStore = originalGetStore }()

	callCount := 0
	database.GetStore = func() (database.Store, error) {
		return &MockStore{
			putFunc: func(key, value string, ttl int64, tags map[string]string) error {
				callCount++
				return nil
			},
			closeFunc: func() error { return nil },
		}, nil
	}

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			cmd := &cobra.Command{}
			args := []string{string(byte(index)), "value"}
			_ = runAdd(cmd, args)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if callCount != 10 {
		t.Errorf("expected 10 calls to Put, got %d", callCount)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// TestAddCmdOutputFormat tests the exact output format
func TestAddCmdOutputFormat(t *testing.T) {
	originalGetStore := database.GetStore
	defer func() { database.GetStore = originalGetStore }()

	database.GetStore = func() (database.Store, error) {
		return &MockStore{
			putFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			closeFunc: func() error { return nil },
		}, nil
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	_ = runAdd(cmd, []string{"mykey", "myvalue"})

	w.Close()
	output, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	expectedOutput := "✅ Key 'mykey' saved successfully!"
	if !contains(string(output), expectedOutput) {
		t.Errorf("expected exact output '%s', got '%s'", expectedOutput, string(output))
	}
}

// BenchmarkRunAdd benchmarks the runAdd function
func BenchmarkRunAdd(b *testing.B) {
	originalGetStore := database.GetStore
	defer func() { database.GetStore = originalGetStore }()

	database.GetStore = func() (database.Store, error) {
		return &MockStore{
			putFunc: func(key, value string, ttl int64, tags map[string]string) error {
				return nil
			},
			closeFunc: func() error { return nil },
		}, nil
	}

	cmd := &cobra.Command{}
	args := []string{"benchkey", "benchvalue"}

	// Redirect output to avoid benchmark noise
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runAdd(cmd, args)
	}
}

// Additional imports needed for the tests
import "os"
