package must

import (
	"fmt"
	"os"
)

const FAILURE = 1

func Must(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(FAILURE)
	}
}
