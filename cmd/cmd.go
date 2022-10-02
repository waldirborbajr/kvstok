package cmd

import (
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/pkg/database"
	"github.com/waldirborbajr/kvstok/pkg/kvpath"
	"github.com/xujiajun/nutsdb"
)

var (
	//	db *nutsdb.DB

	RootCmd = &cobra.Command{
		Use:     "kvstok",
		Short:   "KVStoK is a CLI-based Key Value storage.",
		Version: "0.2.0",
		Long: `KVStoK is an open source software built-in with the main aim of being a
		personal [KEY][VALUE] store, to keep system variables as parameters or passwords
		or anything else stored in a single place.`,
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
