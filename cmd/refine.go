package cmd

import (
	"fmt"
	"strconv"

	"github.com/mauricioTechDev/propcheck-ai/internal/reflection"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var refineCmd = &cobra.Command{
	Use:   "refine",
	Short: "Show or manage refinement reflections",
	Long:  "Show reflection progress or answer individual reflection questions.",
	Example: `  propcheck-ai refine
  propcheck-ai refine reflect 1 --answer "..."
  propcheck-ai refine status`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.Phase != types.PhaseRefine {
			return fmt.Errorf("not in refine phase (current: %s)", s.Phase)
		}

		if len(s.Reflections) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No reflection questions loaded.")
			return nil
		}

		answered := 0
		for _, r := range s.Reflections {
			if r.Answer != "" {
				answered++
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Reflections: %d/%d answered\n", answered, len(s.Reflections))
		return nil
	},
}

var refineReflectAnswerFlag string

var refineReflectCmd = &cobra.Command{
	Use:     "reflect <id>",
	Short:   "Answer a reflection question",
	Long:    "Provide an answer to a specific reflection question by its ID.",
	Example: `  propcheck-ai refine reflect 1 --answer "The generators cover edge cases including empty and single-element inputs"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.Phase != types.PhaseRefine {
			return fmt.Errorf("not in refine phase (current: %s)", s.Phase)
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("reflection ID must be a number, got %q", args[0])
		}

		if refineReflectAnswerFlag == "" {
			return fmt.Errorf("--answer is required")
		}

		if err := reflection.ValidateAnswer(refineReflectAnswerFlag); err != nil {
			return err
		}

		if err := s.AnswerReflection(id, refineReflectAnswerFlag); err != nil {
			return err
		}

		s.AddEvent("reflection_answered", func(e *types.Event) {
			e.Result = fmt.Sprintf("q%d", id)
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}

		pending := s.PendingReflections()
		fmt.Fprintf(cmd.OutOrStdout(), "Reflection [%d] answered. %d remaining.\n", id, len(pending))
		return nil
	},
}

var refineStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all reflection questions with status",
	Long:  "Display all reflection questions and their answered/pending status.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.Phase != types.PhaseRefine {
			return fmt.Errorf("not in refine phase (current: %s)", s.Phase)
		}

		if len(s.Reflections) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No reflection questions loaded.")
			return nil
		}

		for _, r := range s.Reflections {
			status := "pending"
			if r.Answer != "" {
				status = "answered"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  [%d] (%s) %s\n", r.ID, status, r.Question)
			if r.Answer != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "      -> %q\n", r.Answer)
			}
		}
		return nil
	},
}

func init() {
	refineReflectCmd.Flags().StringVar(&refineReflectAnswerFlag, "answer", "", "reflection answer (min 5 words)")
	refineCmd.AddCommand(refineReflectCmd)
	refineCmd.AddCommand(refineStatusCmd)
	rootCmd.AddCommand(refineCmd)
}
