package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// ExpCmd represents the export command
var ExpCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export all keys to a file.",
	Long:    ``,
	Aliases: []string{"exportkv", "e"},
	Run: func(cmd *cobra.Command, args []string) {
		content := make(map[string]string)
		err := database.DB.View(
			func(tx *nutsdb.Tx) error {
				if keys, values, err := tx.GetAll(database.Bucket); err != nil {
					return err
				} else {
					n := len(keys)
					for i := 0; i < n; i++ {
						content[string(keys[i])] = string(values[i])
						fmt.Println(string(keys[i]), " ", string(values[i]))
					}
				}

				configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
				configHash := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.hash"

				// save to file
				fileContent, _ := json.MarshalIndent(content, "", " ")
				_ = os.WriteFile(configFile, fileContent, 0600)

				hash := kvpath.GenHash(configFile)

				_ = os.WriteFile(configHash, []byte(hash), 0600)

				return nil
			})

		must.Must(err, "ExpCmd() - oops! Huston, we have a problem exporting keys.")

		fmt.Printf("Keys exported to ~/.config/kvstok \n Please keep [.json and .hash] files it into safety place.")
	},
}
