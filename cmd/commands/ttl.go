package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// AddCmd represents the addkv command
var TtlCmd = &cobra.Command{
	Use:     "{t}tladdkv [KEY] [VALUE] [TIME_TO_LIVE_IN_MINUTES]",
	Short:   "Add a key with time to be live. Default 1 minute.",
	Long:    ``,
	Aliases: []string{"t"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("addkv requires at least two parameters [key] and [value] the param [ttl] it is optional, the default value it is 1 minute. Please try it again")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.DB.Update(
			func(tx *nutsdb.Tx) error {
				key := []byte(args[0])
				val := []byte(args[1])
				ttl := uint32(60)

				if len(args) == 3 {
					temp_ttl, err := strconv.ParseUint(string([]byte(args[2])), 10, 32)
					must.Must(err, "Third param must be a number.")
					ttl = uint32(temp_ttl) * 60
				}

				return tx.Put(database.Bucket, key, val, ttl)
			}); err != nil {
			fmt.Printf("Error saving value: %s\n", err.Error())
		}
	},
}
