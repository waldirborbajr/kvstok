// cmd/commands/search.go
package commands

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nutsdb/nutsdb"
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

		database.DB.View(func(tx *nutsdb.Tx) error {
			keys, values, _ := tx.GetAll(database.Bucket)
			results := make(map[string]string)

			for i := 0; i < len(keys); i++ {
				k := string(keys[i])
				if matchesPattern(k, pattern, regex, prefix) {
					results[k] = string(values[i])
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
			return nil
		})
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
