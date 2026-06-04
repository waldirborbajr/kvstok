// Package must provides utilities for fatal error handling.
package must

import (
	"fmt"
	"os"
)

// FAILURE is the exit code used when a fatal error occurs.
const FAILURE = 1

// Must checks if err is non-nil and exits the program with a fatal error message.
func Must(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "SYSTEM ERROR: %s \n", message)
		fmt.Fprintf(os.Stderr, "DEBUG  ERROR: %s \n", err.Error())
		os.Exit(FAILURE)
	}
}
