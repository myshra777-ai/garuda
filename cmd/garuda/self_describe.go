package main

import (
	"context"
	"os"

	"github.com/myshra777-ai/garuda/internal/selfdescribe"
	"github.com/spf13/cobra"
)

var selfDescribeOpts = &selfdescribe.Options{}

var selfDescribeCmd = &cobra.Command{
	Use:   "self-describe [path]",
	Short: "Generate a structured, evidence-backed product description",
	Long: `Self-describe extracts capabilities, CLI commands, semantic model,
trust evidence, and roadmap from the codebase, and outputs a structured
JSON or Markdown description.

Examples:
  garuda self describe
  garuda self describe --workspace my-workspace --output product.json
  garuda self describe --markdown --output README.md
  garuda self describe ./my-service --workspace my-workspace
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		selfDescribeOpts.Path = path
		selfDescribeOpts.TenantID = getTenantIDString()
		selfDescribeOpts.DatabaseURL = getDBURL()

		if err := selfdescribe.Run(context.Background(), selfDescribeOpts); err != nil {
			_, _ = os.Stderr.WriteString("❌ " + err.Error() + "\n")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(selfDescribeCmd)
	selfDescribeCmd.Flags().StringVarP(&selfDescribeOpts.Workspace, "workspace", "w", "", "Workspace name for multi-repo aggregation")
	selfDescribeCmd.Flags().StringVarP(&selfDescribeOpts.OutputFile, "output", "o", "", "Write output to file")
	selfDescribeCmd.Flags().BoolVar(&selfDescribeOpts.Markdown, "markdown", false, "Output as README.md skeleton instead of JSON")
}
