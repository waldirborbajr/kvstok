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

// ImpCmd represents the import command
var ImpCmd = &cobra.Command{
	Use:     "import",
	Short:   "Restore all keys from kvstok.json.",
	Long:    ``,
	Aliases: []string{"importkv", "i"},
	Run: func(cmd *cobra.Command, args []string) {
		var dataResult map[string]string

		configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
		configHash := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.hash"

		// Check export file integrity
		file, err := os.ReadFile(configHash)
		must.Must(err, "ImpCmd() - oops! Huston, we have a problem importing keys.")

		currentHash := kvpath.GenHash(configFile)
		storedHash := []byte(file)

		// areEquals := isEquals(currentHash, string(storedHash))
		areEquals := currentHash == string(storedHash)

		if !areEquals {

			fmt.Fprintf(os.Stderr, "JSON export key corrupted. Hashcode are not the same.")
			os.Exit(1)
		}

		if areEquals {
			// Import JSON after integrity check
			file, err = os.ReadFile(configFile)
			must.Must(err, "ImpCmd() - oops! Huston, we have a problem integrity broken.")

			err = json.Unmarshal([]byte(file), &dataResult)
			must.Must(err, "ImpCmd() - failed to parse JSON file")

			store, err := database.GetStore()
			must.Must(err, "ImpCmd() - failed to open store")

			for key, value := range dataResult {
				must.Must(store.Put(key, value, 0, nil), "ImpCmd() - oops! Huston, we have a problem importing keys")
			}
		}

		fmt.Printf("Keys imported successfully.")
	},
}

// func isEquals(param1 string, param2 string) bool {
// 	bret := true

// 	if param1 != param2 {
// 		bret = false
// 	}

// 	return bret
// }
