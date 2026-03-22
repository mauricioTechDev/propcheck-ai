package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/phase"
	"github.com/mauricioTechDev/propcheck-ai/internal/reflection"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var phaseCmd = &cobra.Command{
	Use:   "phase",
	Short: "Show or manage the current PBT phase",
	Long:  "Show the current PBT phase, advance to the next phase, or manually set a phase.",
	Example: `  propcheck-ai phase
  propcheck-ai phase next
  propcheck-ai phase set generate`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), s.Phase)
		return nil
	},
}

var testResultFlag string

var phaseNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Advance to the next PBT phase",
	Long: `Advance to the next phase in the PBT cycle.

Use --test-result to provide the test outcome for phases that require it.`,
	Example: `  propcheck-ai phase next
  propcheck-ai phase next --test-result pass
  propcheck-ai phase next --test-result fail`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		current := s.Phase

		// Phase-specific pre-checks
		switch current {
		case types.PhaseProperty:
			if len(s.ActiveProperties()) == 0 {
				return fmt.Errorf("cannot advance: no active properties")
			}
			if s.CurrentPropertyID == nil {
				return fmt.Errorf("cannot advance: no property selected")
			}

		case types.PhaseGenerate:
			// No specific requirements

		case types.PhaseValidate:
			// Test result is REQUIRED (determines path)
			effectiveResult := resolveTestResult(cmd, s)
			if effectiveResult == "" {
				return fmt.Errorf("cannot advance: test result required for VALIDATE (run 'propcheck-ai test' first)")
			}
			if effectiveResult == "error" {
				return fmt.Errorf("cannot advance: last test run was an infrastructure/environment error. Fix the environment and re-run 'propcheck-ai test'")
			}
			if effectiveResult != "pass" && effectiveResult != "fail" {
				return fmt.Errorf("--test-result must be 'pass' or 'fail', got %q", effectiveResult)
			}

		case types.PhaseShrink:
			if s.ShrinkAnalysis == "" {
				return fmt.Errorf("cannot advance: no shrink analysis recorded")
			}

		case types.PhaseRefine:
			effectiveResult := resolveTestResult(cmd, s)
			if effectiveResult != "" {
				if effectiveResult == "error" {
					return fmt.Errorf("cannot advance: last test run was an infrastructure/environment error. Fix the environment and re-run 'propcheck-ai test'")
				}
				if effectiveResult != "pass" {
					return fmt.Errorf("cannot advance: REFINE phase expects tests to pass, but got test result %s", effectiveResult)
				}
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: advancing without test result. The REFINE phase expects tests to pass.\n")
			}

			if len(s.Reflections) > 0 && !s.AllReflectionsAnswered() {
				pending := s.PendingReflections()
				return fmt.Errorf("cannot advance: %d reflection question(s) unanswered", len(pending))
			}

		case types.PhaseDone:
			return fmt.Errorf("cannot advance past done")
		}

		// Auto-complete current property when leaving REFINE
		if current == types.PhaseRefine && s.CurrentPropertyID != nil {
			completedID := *s.CurrentPropertyID
			if err := s.CompleteCurrentProperty(); err != nil {
				return fmt.Errorf("completing current property: %w", err)
			}
			s.Iteration++
			s.ShrinkAnalysis = ""
			fmt.Fprintf(cmd.OutOrStdout(), "Completed property [%d], iteration %d done\n", completedID, s.Iteration)
		}

		// Determine effective test result for path decisions
		effectiveResult := resolveTestResult(cmd, s)

		// Clear last test result after consuming
		if s.LastTestResult != "" {
			s.LastTestResult = ""
		}

		// Determine next phase
		hasRemaining := len(s.RemainingProperties()) > 0
		next, err := phase.NextInLoop(current, effectiveResult, hasRemaining)
		if err != nil {
			return err
		}

		s.Phase = next
		if next == types.PhaseRefine {
			s.Reflections = reflection.DefaultQuestions()
		}
		if next == types.PhaseProperty {
			s.CurrentPropertyID = nil
		}
		s.AddEvent("phase_next", func(e *types.Event) {
			e.From = string(current)
			e.To = string(next)
			if effectiveResult != "" {
				e.Result = effectiveResult
			}
		})
		if err := session.Save(dir, s); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Phase: %s -> %s\n", current, next)

		if next == types.PhaseProperty && hasRemaining {
			remaining := s.ActiveProperties()
			fmt.Fprintf(cmd.OutOrStdout(), "%d property(ies) remaining\n", len(remaining))
			for _, p := range remaining {
				fmt.Fprintf(cmd.OutOrStdout(), "  [%d] %s\n", p.ID, p.Description)
			}
		}
		return nil
	},
}

// resolveTestResult returns the effective test result from flag or session.
func resolveTestResult(cmd *cobra.Command, s *types.Session) string {
	if testResultFlag != "" {
		return testResultFlag
	}
	if s.LastTestResult != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Using last test result from session: %s\n", s.LastTestResult)
		return s.LastTestResult
	}
	return ""
}

var phaseSetForceFlag bool

var phaseSetCmd = &cobra.Command{
	Use:   "set <property|generate|validate|shrink|refine|done>",
	Short: "Manually set the PBT phase (requires --force)",
	Long: `Override the current phase. Requires --force because this bypasses PBT guardrails.
Prefer 'propcheck-ai phase next' for normal phase advancement.`,
	Example: `  propcheck-ai phase set generate --force
  propcheck-ai phase set validate --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		p := types.Phase(args[0])
		if !p.IsValid() {
			return fmt.Errorf("invalid phase %q. Valid phases: property, generate, validate, shrink, refine, done", args[0])
		}

		if s.AgentMode {
			return fmt.Errorf("phase set is disabled in agent mode. Use 'propcheck-ai phase next' for phase advancement")
		}

		if !phaseSetForceFlag {
			return fmt.Errorf("phase set bypasses PBT guardrails; use --force to override, or prefer 'propcheck-ai phase next'")
		}

		old := s.Phase
		s.Phase = p
		if p == types.PhaseRefine && len(s.Reflections) == 0 {
			s.Reflections = reflection.DefaultQuestions()
		}
		if p == types.PhaseProperty {
			s.CurrentPropertyID = nil
		}
		s.AddEvent("phase_set", func(e *types.Event) {
			e.From = string(old)
			e.To = string(p)
			e.Result = "forced_override"
		})
		if err := session.Save(dir, s); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Phase set to: %s\n", p)
		return nil
	},
}

func init() {
	phaseNextCmd.Flags().StringVar(&testResultFlag, "test-result", "", "test outcome: 'pass' or 'fail'")
	phaseSetCmd.Flags().BoolVar(&phaseSetForceFlag, "force", false, "override PBT guardrails and force phase change")
	phaseCmd.AddCommand(phaseNextCmd)
	phaseCmd.AddCommand(phaseSetCmd)
	rootCmd.AddCommand(phaseCmd)
}
