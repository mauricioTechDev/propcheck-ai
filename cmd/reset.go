package cmd

import (
	"fmt"
	"os"

	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear the PBT session and start over",
	Long:  "Remove the session file so you can reinitialize with 'propcheck-ai init'.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()

		if !session.Exists(dir) {
			return fmt.Errorf("no PBT session to reset")
		}

		if err := os.Remove(session.FilePath(dir)); err != nil {
			return fmt.Errorf("removing session file: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "PBT session reset.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
