package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var testSummaryFlag bool

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the configured test command",
	Long:  "Execute the test command and classify the result as pass, fail, or error.",
	Example: `  propcheck-ai test
  propcheck-ai test --summary`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.TestCmd == "" {
			return fmt.Errorf("no test command configured. Use 'propcheck-ai init --test-cmd \"...\"'")
		}

		c := exec.Command("sh", "-c", s.TestCmd)
		var stdout, stderr bytes.Buffer
		c.Stdout = &stdout
		c.Stderr = &stderr

		runErr := c.Run()

		combined := stdout.String() + stderr.String()
		result := classifyResult(runErr, combined)

		s.LastTestResult = result
		s.AddEvent("test_run", func(e *types.Event) {
			e.Result = result
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}

		if testSummaryFlag {
			lines := strings.Split(strings.TrimRight(combined, "\n"), "\n")
			start := 0
			if len(lines) > 20 {
				start = len(lines) - 20
			}
			for _, line := range lines[start:] {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
		} else {
			fmt.Fprint(cmd.OutOrStdout(), combined)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nTest result: %s\n", result)
		return nil
	},
}

// classifyResult determines if the test run was pass, fail, or error.
func classifyResult(err error, output string) string {
	if err == nil {
		return "pass"
	}
	// Check for infrastructure/environment errors
	lower := strings.ToLower(output)
	errorIndicators := []string{
		"command not found",
		"no such file or directory",
		"permission denied",
		"cannot find",
		"compilation failed",
		"build failed",
		"syntax error",
	}
	for _, indicator := range errorIndicators {
		if strings.Contains(lower, indicator) {
			return "error"
		}
	}
	return "fail"
}

func init() {
	testCmd.Flags().BoolVar(&testSummaryFlag, "summary", false, "show only last 20 lines of output")
	rootCmd.AddCommand(testCmd)
}
