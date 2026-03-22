package phase

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

// standardTransitions: when VALIDATE tests pass, skip SHRINK.
var standardTransitions = map[types.Phase]types.Phase{
	types.PhaseProperty: types.PhaseGenerate,
	types.PhaseGenerate: types.PhaseValidate,
	types.PhaseValidate: types.PhaseRefine,
	types.PhaseRefine:   types.PhaseDone,
}

// failureTransitions: when VALIDATE tests fail, go to SHRINK.
var failureTransitions = map[types.Phase]types.Phase{
	types.PhaseProperty: types.PhaseGenerate,
	types.PhaseGenerate: types.PhaseValidate,
	types.PhaseValidate: types.PhaseShrink,
	types.PhaseShrink:   types.PhaseRefine,
	types.PhaseRefine:   types.PhaseDone,
}

// Next returns the next phase using the standard flow (tests passed).
func Next(current types.Phase) (types.Phase, error) {
	next, ok := standardTransitions[current]
	if !ok {
		return "", fmt.Errorf("no transition from %q", current)
	}
	return next, nil
}

// NextWithResult returns the next phase based on the test result.
// When current is VALIDATE and testResult is "fail", uses failure path (-> SHRINK).
// When current is VALIDATE and testResult is "pass", uses standard path (-> REFINE).
// For all other phases, testResult is ignored and standard transitions apply.
func NextWithResult(current types.Phase, testResult string) (types.Phase, error) {
	if current == types.PhaseValidate && testResult == "fail" {
		return types.PhaseShrink, nil
	}
	if current == types.PhaseShrink {
		return types.PhaseRefine, nil
	}
	next, ok := standardTransitions[current]
	if !ok {
		return "", fmt.Errorf("no transition from %q", current)
	}
	return next, nil
}

// NextInLoop handles the per-property loop.
// At REFINE: returns PROPERTY if properties remain, DONE if empty.
func NextInLoop(current types.Phase, testResult string, hasRemainingProperties bool) (types.Phase, error) {
	if current == types.PhaseRefine {
		if hasRemainingProperties {
			return types.PhaseProperty, nil
		}
		return types.PhaseDone, nil
	}
	return NextWithResult(current, testResult)
}

// ExpectedTestResult returns what test outcome is expected for the given phase.
func ExpectedTestResult(p types.Phase) string {
	switch p {
	case types.PhaseProperty, types.PhaseGenerate:
		return ""
	case types.PhaseValidate:
		return "any"
	case types.PhaseShrink:
		return "fail"
	case types.PhaseRefine:
		return "pass"
	default:
		return ""
	}
}

// CanTransition checks whether moving from one phase to another is valid.
func CanTransition(from, to types.Phase) bool {
	// Check standard transitions
	if next, ok := standardTransitions[from]; ok && next == to {
		return true
	}
	// Check failure transitions
	if next, ok := failureTransitions[from]; ok && next == to {
		return true
	}
	// Allow refine -> property (loop)
	if from == types.PhaseRefine && to == types.PhaseProperty {
		return true
	}
	return false
}
