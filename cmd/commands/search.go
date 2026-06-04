// cmd/commands/search.go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var SearchCmd = &cobra.Command{
	Use:     "search [PATTERN]",
	Short:   "Search for keys matching a pattern (regex or glob).",
	Aliases: []string{"s"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := args[0]
		regex, err := cmd.Flags().GetBool("regex")
		if err != nil {
			fmt.Fprintf(os.Stderr, "SearchCmd() - invalid flag: %v\n", err)
			return
		}

		prefix, err := cmd.Flags().GetBool("prefix")
		if err != nil {
			fmt.Fprintf(os.Stderr, "SearchCmd() - invalid flag: %v\n", err)
			return
		}

		jsonOut, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "SearchCmd() - invalid flag: %v\n", err)
			return
		}

		store, err := database.GetStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "SearchCmd() - failed to open store: %v\n", err)
			return
		}

		entries, err := store.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "SearchCmd() - failed to list entries: %v\n", err)
			return
		}

		results := make(map[string]string)
		for k, e := range entries {
			if matchesPattern(k, pattern, regex, prefix) {
				results[k] = e.Value
			}
		}

		if jsonOut {
			data, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "SearchCmd() - failed to marshal JSON: %v\n", err)
				return
			}
			fmt.Printf("%s\n", data)
		} else {
			for k, v := range results {
				fmt.Printf("%s\t%s\n", k, v)
			}
		}
	},
}

func matchesPattern(key string, pattern string, useRegex bool, prefixOnly bool) bool {
	if useRegex {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return r.MatchString(key)
	}

	if prefixOnly {
		return strings.HasPrefix(key, pattern)
	}

	// Glob pattern
	matched, err := filepath.Match(pattern, key)
	if err != nil {
		return false
	}
	return matched
}

func init() {
	SearchCmd.Flags().Bool("regex", false, "Use regex pattern (default: glob)")
	SearchCmd.Flags().Bool("prefix", false, "Match only at start")
	SearchCmd.Flags().Bool("json", false, "Output as JSON")
}
