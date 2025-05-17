// cmd/list.go
package cmd

import (
	"github.com/davidkohl/gobelix/asterix"
	"github.com/spf13/cobra"
)

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available ASTERIX categories",
		Long: `Display information about available ASTERIX categories and their versions.
This command lists all the ASTERIX categories implemented in the gobelix library.`,
		Run: runList,
	}

	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	// Configure logging
	logger := ConfigureLogger(Verbose, JsonLogs)

	logger.Info("Available ASTERIX Categories")

	// Get information about known categories
	categories := []asterix.Category{
		asterix.Cat021,
		asterix.Cat048,
		asterix.Cat062,
		asterix.Cat063,
		asterix.Cat065,
		asterix.Cat247,
		asterix.Cat252,
	}

	for _, cat := range categories {
		info := asterix.GetCategoryInfo(cat)
		logger.Info("Category",
			"name", info.Name,
			"version", info.Version,
			"description", info.Description,
			"blockable", info.Blockable,
		)
	}
}
