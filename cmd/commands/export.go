package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// LstCmd represents the lstkv command
var ExpCmd = &cobra.Command{
	Use:     "{e}xportkv",
	Short:   "Export all keys to a file.",
	Long:    ``,
	Aliases: []string{"e"},
	Run: func(cmd *cobra.Command, args []string) {
		content := make(map[string]string)
		err := database.DB.View(
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
				_ = ioutil.WriteFile(configFile, fileContent, 0600)

				hash := kvpath.GenHash(configFile)

				_ = ioutil.WriteFile(configHash, []byte(hash), 0600)

				return nil
			})

		must.Must(err, "ExpCmd() - oops! Huston, we have a problem exporting keys.")

		fmt.Printf("Keys exported to ~/.config/kvstok \n Please keep [.json and .hash] files it into safety place.")
	},
}
