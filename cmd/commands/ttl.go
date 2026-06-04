package commands

import (
	"errors"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
	"github.com/waldirborbajr/kvstok/internal/must"
)

// TtlCmd represents the ttl command
var TtlCmd = &cobra.Command{
	Use:     "ttl [KEY] [VALUE] [TIME_TO_LIVE_IN_MINUTES]",
	Short:   "Add a key with time to be live. Default 1 minute.",
	Long:    ``,
	Aliases: []string{"ttladdkv", "t"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("addkv requires at least two parameters [key] and [value] the param [ttl] it is optional, the default value it is 1 minute. Please try it again")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		store, err := database.GetStore()
		must.Must(err, "TTLCmd() - failed to open store")

		ttl := uint32(60)
		if len(args) == 3 {
			temp_ttl, err := strconv.ParseUint(string([]byte(args[2])), 10, 32)
			must.Must(err, "Third param must be a number.")
			ttl = uint32(temp_ttl) * 60
		}

		must.Must(store.Put(args[0], args[1], ttl, nil), "TTLCmd() - oops! Huston, we have a problem adding/updating keys.")
	},
}
