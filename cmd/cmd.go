package cmd

import (
	"errors"
	"fmt"

	"github.com/waldirborbajr/kvstok/pkg/kvpath"

	"github.com/spf13/cobra"
	"github.com/xujiajun/nutsdb"
)

const (
	dbName = ".6B7673" // -> .kvs
	bucket = "kvstok"
)

var (
	db *nutsdb.DB

	RootCmd = &cobra.Command{
		Use:     "kvstok",
		Short:   "KVStoK it is a simple Key Value storage.",
		Version: "0.2.1",
		Long: `KVStoK is an open source software built-in with the main aim of being a
		personal [KEY][VALUE] store, to keep system variables as parameters or passwords
		or anything else stored in a single place.`,
		// Args:    cobra.NoArgs,
	}

	addCmd = &cobra.Command{
		Use:   "addkv [KEY] [VALUE]",
		Short: "Add or Update a value for a key.",
		Long:  "Add or Update a value for a key, be careful using this to avoid lose any information stored",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("addkv requires two parameters [key] and [value]. Please try it again")
			}
			return nil
		},
		Run: addVal,
	}

	getCmd = &cobra.Command{
		Use:   "getkv key",
		Short: "Get a value for a key.",
		Long:  "Get a value for a key previously stored into database.",
		Args:  cobra.MinimumNArgs(1),
		Run:   getVal,
	}

	deleteCmd = &cobra.Command{
		Use:   "delkv key",
		Short: "Remove a key.",
		Long:  "Remove a key previously stored into database.",
		Args:  cobra.MinimumNArgs(1),
		Run:   deleteVal,
	}

	listCmd = &cobra.Command{
		Use:   "listkv [(-o|--output=)json|yaml]",
		Short: "List all keys values pairs.",
		Long:  "List all keys values pairs stored into database, you can export to file too informing [output] option.",
		Run:   listVal,
	}
)

func init() {
	RootCmd.CompletionOptions.HiddenDefaultCmd = true
	RootCmd.DisableSuggestions = true

	RootCmd.AddCommand(
		addCmd,
		getCmd,
		deleteCmd,
		listCmd,
	)

	homePath := kvpath.GetKVHomeDir() + "/" + dbName

	db, _ = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(homePath),
		nutsdb.WithSegmentSize(2048*2048),
	)
}

func addVal(cmd *cobra.Command, args []string) {
	if err := db.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(args[0])
			val := []byte(args[1])
			return tx.Put(bucket, key, val, 0)
		}); err != nil {
		fmt.Printf("Error saving value: %s\n", err.Error())
	}
}

func getVal(cmd *cobra.Command, args []string) {
	if err := db.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(args[0])
			content, err := tx.Get(bucket, key)
			if err != nil {
				fmt.Printf("Error getting value: Key [%s] does not exists \n", string(key))
			}
			fmt.Printf("%s\n", content.Value)
			return nil
		}); err != nil {
	}
}

func deleteVal(cmd *cobra.Command, args []string) {
	if err := db.Update(
		func(tx *nutsdb.Tx) error {
			key := []byte(args[0])
			return tx.Delete(bucket, key)
		}); err != nil {
		fmt.Printf("Error deleting value: %s\n", err.Error())
	}
}

func listVal(cmd *cobra.Command, args []string) {
	if err := db.View(
		func(tx *nutsdb.Tx) error {
			if nodes, err := tx.GetAll(bucket); err != nil {
				return err
			} else {
				for _, node := range nodes {
					fmt.Println(string(node.Key), " ", string(node.Value))
				}
			}

			return nil
		}); err != nil {
		fmt.Printf("Error listing keys database keys must be empty: %s", err.Error())
	}
}
