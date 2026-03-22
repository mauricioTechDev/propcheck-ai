package cmd

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var (
	initTestCmdFlag string
	initAgentFlag   bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new PBT session",
	Long:  "Create a new property-based testing session in the current directory.",
	Example: `  propcheck-ai init
  propcheck-ai init --test-cmd "go test ./..."
  propcheck-ai init --test-cmd "npm test" --agent`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()

		if session.Exists(dir) {
			return fmt.Errorf("PBT session already exists. Use 'propcheck-ai reset' to start over")
		}

		s, err := session.Create(dir)
		if err != nil {
			return err
		}

		if initTestCmdFlag != "" {
			s.TestCmd = initTestCmdFlag
		}
		if initAgentFlag {
			s.AgentMode = true
		}

		s.AddEvent("init", func(e *types.Event) {
			if initAgentFlag {
				e.Result = "agent_mode"
			}
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "PBT session initialized.")
		if initAgentFlag {
			fmt.Fprintln(cmd.OutOrStdout(), "Agent mode enabled. phase set is disabled.")
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initTestCmdFlag, "test-cmd", "", "test command to run (e.g., 'go test ./...')")
	initCmd.Flags().BoolVar(&initAgentFlag, "agent", false, "enable agent mode (stricter enforcement)")
	rootCmd.AddCommand(initCmd)
}
