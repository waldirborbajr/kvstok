package must

import (
	"fmt"
	"os"
)

func Must(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
