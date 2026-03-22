package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var completeForceFlag bool

var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Finish the PBT cycle",
	Long: `Complete the PBT session by advancing to done and marking all active properties as completed.
In agent mode, requires --force.`,
	Example: `  propcheck-ai complete --test-result pass
  propcheck-ai complete --force`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.AgentMode && !completeForceFlag {
			return fmt.Errorf("complete bypasses PBT guardrails in agent mode. Use --force to override")
		}

		if s.Phase == types.PhaseDone {
			return fmt.Errorf("already in done phase")
		}

		// Validate test result if available
		if s.LastTestResult != "" && s.LastTestResult != "pass" {
			return fmt.Errorf("cannot complete: tests must pass (got %s)", s.LastTestResult)
		}

		count := s.CompleteAllProperties()
		s.Phase = types.PhaseDone
		s.ShrinkAnalysis = ""
		s.CurrentPropertyID = nil

		s.AddEvent("complete", func(e *types.Event) {
			e.PropCount = count
			if completeForceFlag {
				e.Result = "forced"
			}
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "PBT session complete. %d property(ies) finished.\n", count)
		return nil
	},
}

func init() {
	completeCmd.Flags().BoolVar(&completeForceFlag, "force", false, "force completion (required in agent mode)")
	rootCmd.AddCommand(completeCmd)
}
