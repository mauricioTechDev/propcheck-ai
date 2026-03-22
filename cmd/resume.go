package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/formatter"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Show compact session checkpoint for context recovery",
	Long:  "Display a compact summary of the current session state for agent re-orientation.",
	Example: `  propcheck-ai resume
  propcheck-ai resume --format json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		out, err := formatter.FormatResume(s, formatter.Format(formatFlag))
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
