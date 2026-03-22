package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	version    = "dev"
	formatFlag string
)

var rootCmd = &cobra.Command{
	Use:   "propcheck-ai",
	Short: "PBT guardrails for AI coding agents",
	Long: `propcheck-ai is a property-based testing state machine for AI coding agents.

It provides phase tracking, property management, and phase-gating. The state
machine enforces PBT discipline — constraints are enforced, not instructed.

The CLI does NOT run tests — the AI agent runs tests itself. This tool
provides the state machine and guardrails that keep the PBT loop tight.`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if !cmd.Flags().Changed("format") && !isTerminal() {
			formatFlag = "json"
		}
	},
}

// isTerminal reports whether stdout is connected to a terminal.
var isTerminal = func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "text", "output format: text or json (default: json when non-interactive)")
}

func getWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}
	return dir
}
