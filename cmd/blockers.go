package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/phase"
	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/spf13/cobra"
)

var blockersCmd = &cobra.Command{
	Use:   "blockers",
	Short: "Show what is blocking phase advancement",
	Long:  "Display conditions preventing advancement from the current phase.",
	Example: `  propcheck-ai blockers
  propcheck-ai blockers --format json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		blockers := phase.GetBlockers(s)

		if formatFlag == "json" {
			type blockersOutput struct {
				Phase      string   `json:"phase"`
				Blockers   []string `json:"blockers"`
				CanAdvance bool     `json:"can_advance"`
			}
			out := blockersOutput{
				Phase:      string(s.Phase),
				Blockers:   blockers,
				CanAdvance: len(blockers) == 0,
			}
			if out.Blockers == nil {
				out.Blockers = []string{}
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		if len(blockers) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No blockers. Ready to advance.")
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Blockers:")
		for _, b := range blockers {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", b)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blockersCmd)
}
