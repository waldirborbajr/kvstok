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

// AddCmd represents the addkv command
var ImpCmd = &cobra.Command{
	Use:     "{i}mportkv",
	Short:   "Rostore all keys from kvstok.json.",
	Long:    ``,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		var dataResult map[string]string

		configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
		configHash := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.hash"

		// Check export file integrity
		file, err := os.ReadFile(configHash)
		must.Must(err, "ImpCmd() - oops! Huston, we have a problem importing keys.")

		currentHash := kvpath.GenHash(configFile)
		storedHash := []byte(file)

		areEquals := isEquals(currentHash, string(storedHash))

		if !areEquals {

			fmt.Fprintf(os.Stderr, "JSON export key corrupted. Hashcode are not the same.")
			os.Exit(1)
		}

		if areEquals {
			// Import JSON after integrity check
			file, err = os.ReadFile(configFile)
			must.Must(err, "ImpCmd() - oops! Huston, we have a problem integrity broken.")

			json.Unmarshal([]byte(file), &dataResult)

			for key, value := range dataResult {
				err := database.DB.Update(
					func(tx *nutsdb.Tx) error {
						key := []byte(key)
						val := []byte(value)
						return tx.Put(database.Bucket, key, val, 0)
					})

				must.Must(err, "ImpCmd() - oops! Huston, we have a problem integrity broken.")
			}
		}

		fmt.Printf("Keys imported successfully.")
	},
}

func isEquals(param1 string, param2 string) bool {
	bret := true

	if param1 != param2 {
		bret = false
	}

	return bret
}
