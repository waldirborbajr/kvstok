package main

import (
	"github.com/waldirborbajr/kvstok/cmd"
	"github.com/waldirborbajr/kvstok/pkg/must"
)

func main() {
	must.Must(cmd.RootCmd.Execute())
}
