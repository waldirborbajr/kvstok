package commands

import (
	"fmt"
	"strings"

	"github.com/nutsdb/nutsdb"
	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// cmd/commands/env.go - NOVO

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Export keys as environment variables.",
	Flags: map[string]string{
		"--format": "dotenv|shell|json (default: dotenv)",
	},
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")

		database.DB.View(func(tx *nutsdb.Tx) error {
			keys, values, _ := tx.GetAll(database.Bucket)

			for i := 0; i < len(keys); i++ {
				k := string(keys[i])
				v := string(values[i])

				switch format {
				case "dotenv":
					fmt.Printf("export %s='%s'\n", k, escapeShell(v))
				case "json":
					// JSON output
				case "shell":
					// Shell export
				}
			}
			return nil
		})
	},
}

func escapeShell(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}
