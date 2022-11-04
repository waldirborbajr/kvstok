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

type Kvstok struct {
	Key string
	Val string
}

// AddCmd represents the addkv command
var ImpCmd = &cobra.Command{
	Use:     "importkv",
	Short:   "Rostore all keys from kvstok.json.",
	Aliases: []string{"i"},
	Run:     impVal,
}

func impVal(cmd *cobra.Command, args []string) {

	var dataResult map[string]string

	configFile := kvpath.GetKVHomeDir() + "/.config/kvstok.json"

	fmt.Println(configFile)

	file, err := ioutil.ReadFile(configFile)
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
