package commands

import (
	"encoding/json"
	"fmt"
	"os"

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
		store, err := database.GetStore()
		must.Must(err, "ExpCmd() - failed to open store")

		entries, err := store.List()
		must.Must(err, "ExpCmd() - oops! Huston, we have a problem exporting keys.")

		for k, e := range entries {
			content[k] = e.Value
			fmt.Println(k, " ", e.Value)
		}

		configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
		configHash := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.hash"

		// save to file
		fileContent, _ := json.MarshalIndent(content, "", " ")
		_ = os.WriteFile(configFile, fileContent, 0600)

		hash := kvpath.GenHash(configFile)

		_ = os.WriteFile(configHash, []byte(hash), 0600)

		fmt.Printf("Keys exported to ~/.config/kvstok \n Please keep [.json and .hash] files it into safety place.")
	},
}
