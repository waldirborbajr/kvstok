package must

import (
	"fmt"
	"os"
)

// Fail code
const FAILURE = 1

// Check if error and exit the program
func Must(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(FAILURE)
	}
}
