package kvpath

import (
	"testing"
)

func TestGetKvPath(t *testing.T) {
	path := GetKVPath()

	if path == "" {
		t.Fatal("Error getting path")
	}
}

func TestGetKVHomeDir(t *testing.T) {
	path := GetKVHomeDir()

	if path == "" {
		t.Fatal("Error getting home directory")
	}
}
