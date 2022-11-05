package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/kvpath"
	"github.com/xujiajun/nutsdb"
)

// AddCmd represents the addkv command
var ImpCmd = &cobra.Command{
	Use:     "importkv",
	Short:   "Rostore all keys from kvstok.json.",
	Aliases: []string{"i"},
	Run:     impVal,
}

func isEquals(param1 string, param2 string) bool {
	bret := true

	if param1 != param2 {
		bret = false
	}

	return bret
}

func impVal(cmd *cobra.Command, args []string) {
	var dataResult map[string]string

	configFile := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.json"
	configHash := kvpath.GetKVHomeDir() + "/.config/kvstok/kvstok.hash"

	// Check export file integrity
	file, err := ioutil.ReadFile(configHash)
	if err != nil {
		log.Fatal(err)
	}

	currentHash := kvpath.GenHash(configFile)
	storedHash := []byte(file)

	// fmt.Printf("current: %s \n stored: %s \n", currentHash, storedHash)

	areEquals := isEquals(currentHash, string(storedHash))

	if !areEquals {
		log.Fatal("JSON export key corrupted Hash code are not the same.")
	}

	if areEquals {
		// Import JSON after integrity check
		file, err = ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}

		json.Unmarshal([]byte(file), &dataResult)

		for key, value := range dataResult {
			if err := database.DB.Update(
				func(tx *nutsdb.Tx) error {
					key := []byte(key)
					val := []byte(value)
					return tx.Put(database.Bucket, key, val, 0)
				}); err != nil {
				fmt.Printf("Error saving value: %s\n", err.Error())
			}
		}
	}

}
