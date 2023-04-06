package must

import (
	"fmt"
	"os"
)

// Fail code
const FAILURE = 1

// Check if error and exit the program
func Must(err error, message string) {
	if err != nil {

		fmt.Fprintf(os.Stderr, "SYSTEM ERROR: %s \n", message)
		fmt.Fprintf(os.Stderr, "DEBUG  ERROR: %s \n", err.Error())
		os.Exit(FAILURE)
	}
}
