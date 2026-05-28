package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/waldirborbajr/kvstok/internal/database"
)

var TagCmd = &cobra.Command{
	Use:   "tag [add|remove|list] [KEY] [TAG...]",
	Short: "Manage tags for a key.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("tag requires a subcommand and key")
		}
		subcommand := args[0]
		if subcommand != "add" && subcommand != "remove" && subcommand != "list" {
			return fmt.Errorf("unknown tag command: %s", subcommand)
		}
		return nil
	},
	RunE: runTag,
}

func runTag(cmd *cobra.Command, args []string) error {
	action := args[0]
	key := args[1]
	tags := args[2:]

	store, err := GetStore()
	if err != nil {
		return err
	}
	defer store.Close()

	switch action {
	case "add":
		if len(tags) == 0 {
			return errors.New("please specify at least one tag to add")
		}
		_, entry, err := store.GetRaw(key)
		if err != nil {
			return err
		}
		entry.Tags = append(entry.Tags, tags...)
		entry.Tags = uniqueStrings(entry.Tags)
		return store.Put(key, entry.Value, entry.TTL, entry.Tags)

	case "remove":
		if len(tags) == 0 {
			return errors.New("please specify at least one tag to remove")
		}
		_, entry, err := store.GetRaw(key)
		if err != nil {
			return err
		}
		entry.Tags = removeStrings(entry.Tags, tags)
		return store.Put(key, entry.Value, entry.TTL, entry.Tags)

	case "list":
		_, entry, err := store.GetRaw(key)
		if err != nil {
			return err
		}
		fmt.Printf("Tags for %s: %s\n", key, strings.Join(entry.Tags, ", "))
		return nil

	default:
		return fmt.Errorf("unknown tag command: %s", action)
	}
}

func uniqueStrings(list []string) []string {
	seen := make(map[string]struct{}, len(list))
	result := make([]string, 0, len(list))
	for _, item := range list {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func removeStrings(source []string, targets []string) []string {
	targetMap := make(map[string]struct{}, len(targets))
	for _, tag := range targets {
		targetMap[tag] = struct{}{}
	}
	result := make([]string, 0, len(source))
	for _, tag := range source {
		if _, keep := targetMap[tag]; !keep {
			result = append(result, tag)
		}
	}
	return result
}
