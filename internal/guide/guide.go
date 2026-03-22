package guide

import (
	"github.com/mauricioTechDev/propcheck-ai/internal/phase"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

// Generate produces state-only guidance for the current PBT session.
func Generate(s *types.Session) types.Guidance {
	g := types.Guidance{
		Phase:           s.Phase,
		TestCmd:         s.TestCmd,
		Properties:      s.ActiveProperties(),
		Iteration:       s.Iteration,
		TotalProperties: len(s.Properties),
	}

	if cp := s.CurrentProperty(); cp != nil {
		g.CurrentProperty = cp
	}

	// Compute next phase (use last test result for VALIDATE branching)
	testResult := s.LastTestResult
	if next, err := phase.NextWithResult(s.Phase, testResult); err == nil {
		g.NextPhase = next
	}

	// Expected test result
	if s.Phase != types.PhaseDone {
		g.ExpectedTestResult = phase.ExpectedTestResult(s.Phase)
	}

	// Blockers
	g.Blockers = phase.GetBlockers(s)

	// Include reflections during REFINE phase
	if s.Phase == types.PhaseRefine {
		g.Reflections = s.Reflections
	}

	// Include shrink analysis during SHRINK or REFINE phases
	if s.Phase == types.PhaseShrink || s.Phase == types.PhaseRefine {
		g.ShrinkAnalysis = s.ShrinkAnalysis
	}

	return g
}
