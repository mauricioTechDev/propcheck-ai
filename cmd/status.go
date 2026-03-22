package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/formatter"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show full session overview",
	Long:  "Display phase, properties, compliance score, and event history.",
	Example: `  propcheck-ai status
  propcheck-ai status --format json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		out, err := formatter.FormatFullStatus(s, formatter.Format(formatFlag))
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
