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
		regex, _ := cmd.Flags().GetBool("regex")
		prefix, _ := cmd.Flags().GetBool("prefix")
		jsonOut, _ := cmd.Flags().GetBool("json")

		store, err := database.GetStore()
		mustErr := func(e error, msg string) {
			if e != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", msg, e)
			}
		}

		entries, err := store.List()
		if err != nil {
			mustErr(err, "SearchCmd() - failed to list entries")
			return
		}

		results := make(map[string]string)
		for k, e := range entries {
			if matchesPattern(k, pattern, regex, prefix) {
				results[k] = e.Value
			}
		}

		if jsonOut {
			data, _ := json.MarshalIndent(results, "", "  ")
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
		r, _ := regexp.Compile(pattern)
		return r.MatchString(key)
	} else if prefixOnly {
		return strings.HasPrefix(key, pattern)
	} else {
		// Glob pattern
		matched, _ := filepath.Match(pattern, key)
		return matched
	}
}

func init() {
	SearchCmd.Flags().Bool("regex", false, "Use regex pattern (default: glob)")
	SearchCmd.Flags().Bool("prefix", false, "Match only at start")
	SearchCmd.Flags().Bool("json", false, "Output as JSON")
}
