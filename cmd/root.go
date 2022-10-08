package cmd

import (
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/waldirborbajr/kvstok/pkg/kvpath"
	"github.com/waldirborbajr/kvstok/pkg/must"
	"github.com/xujiajun/nutsdb"
)

const DBSIZE = 2048 * 2048

var (
	RootCmd = &cobra.Command{
		Use:     "kvstok",
		Short:   "KVStoK is a CLI-based KEY VALUE storage.",
		Version: "0.2.0",
	}
)

func Execute() {
	must.Must(RootCmd.Execute())
}

func init() {
	RootCmd.CompletionOptions.HiddenDefaultCmd = true
	RootCmd.DisableSuggestions = true

	homePath := kvpath.GetKVHomeDir() + "/" + database.DBName

	database.DB, _ = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(homePath),
		nutsdb.WithSegmentSize(DBSIZE),
	)
}
