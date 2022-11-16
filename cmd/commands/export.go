package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/xujiajun/nutsdb"
)

// LstCmd represents the lstkv command
var ExpCmd = &cobra.Command{
	Use:     "exportkv",
	Short:   "Export all keys to a file.",
	Aliases: []string{"e"},
	Run:     exportVal,
}

func exportVal(cmd *cobra.Command, args []string) {
	content := make(map[string]string)

	if err := database.DB.View(
		func(tx *nutsdb.Tx) error {
			if nodes, err := tx.GetAll(database.Bucket); err != nil {
				return err
			} else {
				for _, node := range nodes {
					content[string(node.Key)] = string(node.Value)
				}
			}

			configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
			configHash := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.hash"

			// save to file
			fileContent, _ := json.MarshalIndent(content, "", " ")
			_ = ioutil.WriteFile(configFile, fileContent, 0o644)

			hash := kvpath.GenHash(configFile)

			_ = ioutil.WriteFile(configHash, []byte(hash), 0o644)

			return nil
		}); err != nil {
		fmt.Printf("Error listing keys database keys must be empty: %s", err.Error())
	}
}
