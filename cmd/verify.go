package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/verify"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Check PBT compliance",
	Long:  "Analyze session history for PBT violations and report a compliance score.",
	Example: `  propcheck-ai verify
  propcheck-ai verify --format json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		result := verify.Analyze(s)

		if formatFlag == "json" {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
		} else {
			if len(result.Violations) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "PBT Compliance: PASS")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "PBT Compliance: VIOLATIONS FOUND")
				for _, v := range result.Violations {
					fmt.Fprintf(cmd.OutOrStdout(), "  - [%s] %s\n", v.Rule, v.Message)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Score: %.0f%% (%d/%d properties compliant)\n",
				result.Score, result.PropertiesCompliant, result.PropertiesVerified)
		}

		if !result.Compliant {
			// Return error for non-zero exit code
			return fmt.Errorf("PBT compliance check failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
