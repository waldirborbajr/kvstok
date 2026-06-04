package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

// EnvCmd exports all stored keys as environment variables.
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Export keys as environment variables.",
	Run: func(cmd *cobra.Command, _ []string) {
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			fmt.Fprintf(os.Stderr, "EnvCmd() - invalid flag: %v\n", err)
			return
		}

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
			encoded, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "EnvCmd() - failed to marshal JSON: %v\n", err)
				return
			}
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
