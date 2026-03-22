package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/formatter"
	"github.com/mauricioTechDev/propcheck-ai/internal/guide"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/spf13/cobra"
)

var guideCmd = &cobra.Command{
	Use:   "guide",
	Short: "Show current PBT state for AI agents",
	Long:  "Display phase, properties, blockers, and expected test result.",
	Example: `  propcheck-ai guide
  propcheck-ai guide --format json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		g := guide.Generate(s)
		out, err := formatter.FormatGuidance(g, formatter.Format(formatFlag))
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(guideCmd)
}
