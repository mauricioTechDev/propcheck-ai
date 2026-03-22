package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/reflection"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var shrinkCmd = &cobra.Command{
	Use:   "shrink",
	Short: "Show or manage counter-example analysis",
	Long:  "Show the current shrink analysis status or record a counter-example analysis.",
	Example: `  propcheck-ai shrink
  propcheck-ai shrink analyze --answer "The minimal failing input was..."`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.ShrinkAnalysis == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "No counter-example analysis recorded yet.")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Analysis: %s\n", s.ShrinkAnalysis)
		}
		return nil
	},
}

var shrinkAnalyzeAnswerFlag string

var shrinkAnalyzeCmd = &cobra.Command{
	Use:     "analyze",
	Short:   "Record counter-example analysis",
	Long:    "Document the minimal failing input, expected vs actual behavior, and root cause.",
	Example: `  propcheck-ai shrink analyze --answer "The minimal failing input was [-3, 0, 5]. The sort placed 0 after 5 because the comparator used > instead of >=."`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.Phase != types.PhaseShrink {
			return fmt.Errorf("not in shrink phase (current: %s). Analysis is only available during shrink", s.Phase)
		}

		if shrinkAnalyzeAnswerFlag == "" {
			return fmt.Errorf("--answer is required")
		}

		if err := reflection.ValidateAnswer(shrinkAnalyzeAnswerFlag); err != nil {
			return err
		}

		s.ShrinkAnalysis = shrinkAnalyzeAnswerFlag
		s.AddEvent("shrink_analyze", func(e *types.Event) {
			e.Result = "recorded"
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Counter-example analysis recorded.")
		return nil
	},
}

func init() {
	shrinkAnalyzeCmd.Flags().StringVar(&shrinkAnalyzeAnswerFlag, "answer", "", "counter-example analysis text (min 5 words)")
	shrinkCmd.AddCommand(shrinkAnalyzeCmd)
	rootCmd.AddCommand(shrinkCmd)
}
