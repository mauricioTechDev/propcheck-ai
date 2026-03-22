package phase

import (
	"testing"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

func assertContains(t *testing.T, blockers []string, want string) {
	t.Helper()
	for _, b := range blockers {
		if b == want {
			return
		}
	}
	t.Errorf("blockers %v does not contain %q", blockers, want)
}

func assertEmpty(t *testing.T, blockers []string) {
	t.Helper()
	if len(blockers) != 0 {
		t.Errorf("expected no blockers, got %v", blockers)
	}
}

func TestBlockersPropertyNoProperties(t *testing.T) {
	s := types.NewSession()
	blockers := GetBlockers(s)
	assertContains(t, blockers, "No active properties")
}

func TestBlockersPropertyNoSelection(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("test", "")
	blockers := GetBlockers(s)
	assertContains(t, blockers, "No property selected")
}

func TestBlockersPropertyReady(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("test", "")
	s.SetCurrentProperty(1)
	blockers := GetBlockers(s)
	assertEmpty(t, blockers)
}

func TestBlockersGenerateNoBlockers(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseGenerate
	blockers := GetBlockers(s)
	assertEmpty(t, blockers)
}

func TestBlockersValidateNoTestResult(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseValidate
	blockers := GetBlockers(s)
	assertContains(t, blockers, "No test result recorded")
}

func TestBlockersValidateWithPass(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseValidate
	s.LastTestResult = "pass"
	blockers := GetBlockers(s)
	assertEmpty(t, blockers)
}

func TestBlockersValidateWithFail(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseValidate
	s.LastTestResult = "fail"
	blockers := GetBlockers(s)
	// Both pass and fail are valid for VALIDATE
	assertEmpty(t, blockers)
}

func TestBlockersShrinkNoAnalysis(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseShrink
	blockers := GetBlockers(s)
	assertContains(t, blockers, "No shrink analysis recorded")
}

func TestBlockersShrinkWithAnalysis(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseShrink
	s.ShrinkAnalysis = "The minimal failing input was [-3, 0, 5]"
	blockers := GetBlockers(s)
	assertEmpty(t, blockers)
}

func TestBlockersRefineNoTestResult(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseRefine
	blockers := GetBlockers(s)
	assertContains(t, blockers, "No test result recorded")
}

func TestBlockersRefineTestFail(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseRefine
	s.LastTestResult = "fail"
	blockers := GetBlockers(s)
	assertContains(t, blockers, "Test result 'fail' does not match expected 'pass'")
}

func TestBlockersRefineUnansweredReflections(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseRefine
	s.LastTestResult = "pass"
	s.Reflections = []types.ReflectionQuestion{
		{ID: 1, Question: "Q1"},
		{ID: 2, Question: "Q2"},
	}
	blockers := GetBlockers(s)
	assertContains(t, blockers, "2 reflection question(s) unanswered")
}

func TestBlockersRefineAllAnswered(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseRefine
	s.LastTestResult = "pass"
	s.Reflections = []types.ReflectionQuestion{
		{ID: 1, Question: "Q1", Answer: "answered this question here"},
	}
	blockers := GetBlockers(s)
	assertEmpty(t, blockers)
}

func TestBlockersDone(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseDone
	blockers := GetBlockers(s)
	assertContains(t, blockers, "Cannot advance past done")
}
