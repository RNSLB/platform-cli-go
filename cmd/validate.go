package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RNSLB/platform-cli/pkg/namespace"
)

var validateCmd = &cobra.Command{
	Use:   "validate [yaml-file]",
	Short: "Validate namespace configurations from YAML file",
	Long: `Validate namespace configurations from a YAML file.

Checks each namespace for:
- Valid name format (lowercase, alphanumeric, hyphens)
- Name length (1-63 characters)
- No leading/trailing hyphens

Example:
  platform validate namespaces.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		fmt.Printf("📂 Loading namespaces from: %s\n", filename)

		namespaces, err := namespace.LoadFromYAML(filename)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Loaded %d namespaces\n\n", len(namespaces))

		validCount, invalidCount, results := namespace.ValidateAll(namespaces)

		fmt.Println("=== Validation Results ===\n")
		for _, result := range results {
			if result.Valid {
				fmt.Printf("✓ %s (Team: %s, Env: %s)\n",
					result.Namespace.Name,
					result.Namespace.Team,
					result.Namespace.Environment)
			} else {
				fmt.Printf("✗ %s - %s\n",
					result.Namespace.Name,
					result.Reason)
			}
		}

		fmt.Printf("\n=== Summary ===\n")
		fmt.Printf("Total: %d\n", len(namespaces))
		fmt.Printf("Valid: %d ✓\n", validCount)
		fmt.Printf("Invalid: %d ✗\n", invalidCount)

		if invalidCount > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
