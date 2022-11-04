package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"

	"github.com/spf13/cobra"
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

	user, err := user.Current()
	if err != nil {
		log.Fatal("Current ", err)
	}

	home := user.HomeDir
	configFile := home + "/.config/kvstok.json"

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	type Keys struct {
		key string
		val string
	}

	var obj Keys

	err = json.Unmarshal(data, &obj)
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Println(obj)

	// if err := database.DB.Update(
	// 	func(tx *nutsdb.Tx) error {
	// 		key := []byte(args[0])
	// 		val := []byte(args[1])
	// 		return tx.Put(database.Bucket, key, val, 0)
	// 	}); err != nil {
	// 	fmt.Printf("Error saving value: %s\n", err.Error())
	// }
}
