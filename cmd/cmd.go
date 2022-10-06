package cmd

import (
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/waldirborbajr/kvstok/pkg/kvpath"
	"github.com/xujiajun/nutsdb"
)

var (
	RootCmd = &cobra.Command{
		Use:     "kvstok",
		Short:   "KVStoK is a CLI-based Key Value storage.",
		Version: "0.2.0",
	}
)

func init() {
	RootCmd.CompletionOptions.HiddenDefaultCmd = true
	RootCmd.DisableSuggestions = true

	homePath := kvpath.GetKVHomeDir() + "/" + database.DBName

	database.DB, _ = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(homePath),
		nutsdb.WithSegmentSize(2048*2048),
	)
}
