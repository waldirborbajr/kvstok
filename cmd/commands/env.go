package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Export keys as environment variables.",
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")

		store, err := database.GetStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "EnvCmd() - failed to open store: %v\n", err)
			return
		}

		entries, err := store.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "EnvCmd() - failed to list entries: %v\n", err)
			return
		}

		if format == "json" {
			output := make(map[string]string)
			for k, e := range entries {
				output[k] = e.Value
			}
			encoded, _ := json.MarshalIndent(output, "", "  ")
			fmt.Printf("%s\n", encoded)
			return
		}

		for k, e := range entries {
			v := e.Value
			switch format {
			case "dotenv":
				fmt.Printf("export %s='%s'\n", k, escapeShell(v))
			case "shell":
				fmt.Printf("export %s='%s'\n", k, escapeShell(v))
			default:
				fmt.Printf("export %s='%s'\n", k, escapeShell(v))
			}
		}
	},
}

func init() {
	EnvCmd.Flags().String("format", "dotenv", "dotenv|shell|json (default: dotenv)")
}

func escapeShell(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}
