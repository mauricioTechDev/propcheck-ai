package phase

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

// GetBlockers returns conditions preventing advancement from the current phase.
func GetBlockers(s *types.Session) []string {
	var blockers []string

	switch s.Phase {
	case types.PhaseProperty:
		if len(s.ActiveProperties()) == 0 {
			blockers = append(blockers, "No active properties")
		}
		if s.CurrentPropertyID == nil && len(s.ActiveProperties()) > 0 {
			blockers = append(blockers, "No property selected")
		}

	case types.PhaseGenerate:
		// No specific blockers — generators are code, not gated by test results

	case types.PhaseValidate:
		if s.LastTestResult == "" {
			blockers = append(blockers, "No test result recorded")
		}

	case types.PhaseShrink:
		if s.ShrinkAnalysis == "" {
			blockers = append(blockers, "No shrink analysis recorded")
		}

	case types.PhaseRefine:
		if s.LastTestResult == "" {
			blockers = append(blockers, "No test result recorded")
		} else if s.LastTestResult != "pass" {
			blockers = append(blockers,
				fmt.Sprintf("Test result '%s' does not match expected 'pass'", s.LastTestResult))
		}
		if len(s.Reflections) > 0 {
			pending := s.PendingReflections()
			if len(pending) > 0 {
				blockers = append(blockers,
					fmt.Sprintf("%d reflection question(s) unanswered", len(pending)))
			}
		}

	case types.PhaseDone:
		blockers = append(blockers, "Cannot advance past done")
	}

	return blockers
}
