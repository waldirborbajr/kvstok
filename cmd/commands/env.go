package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Export keys as environment variables.",
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")

		database.DB.View(func(tx *nutsdb.Tx) error {
			keys, values, _ := tx.GetAll(database.Bucket)

			if format == "json" {
				output := make(map[string]string)
				for i := 0; i < len(keys); i++ {
					output[string(keys[i])] = string(values[i])
				}
				encoded, _ := json.MarshalIndent(output, "", "  ")
				fmt.Printf("%s\n", encoded)
				return nil
			}

			for i := 0; i < len(keys); i++ {
				k := string(keys[i])
				v := string(values[i])

				switch format {
				case "dotenv":
					fmt.Printf("export %s='%s'\n", k, escapeShell(v))
				case "shell":
					fmt.Printf("export %s='%s'\n", k, escapeShell(v))
				default:
					fmt.Printf("export %s='%s'\n", k, escapeShell(v))
				}
			}
			return nil
		})
	},
}

func init() {
	EnvCmd.Flags().String("format", "dotenv", "dotenv|shell|json (default: dotenv)")
}

func escapeShell(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}
