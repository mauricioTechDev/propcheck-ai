package guide

import (
	"testing"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

func TestGeneratePropertyPhase(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("sort idempotent", "invariant")
	s.SetCurrentProperty(1)

	g := Generate(s)

	if g.Phase != types.PhaseProperty {
		t.Errorf("Phase = %q, want %q", g.Phase, types.PhaseProperty)
	}
	if g.NextPhase != types.PhaseGenerate {
		t.Errorf("NextPhase = %q, want %q", g.NextPhase, types.PhaseGenerate)
	}
	if g.CurrentProperty == nil {
		t.Fatal("CurrentProperty should not be nil")
	}
	if g.CurrentProperty.ID != 1 {
		t.Errorf("CurrentProperty.ID = %d, want 1", g.CurrentProperty.ID)
	}
	if len(g.Properties) != 1 {
		t.Errorf("Properties count = %d, want 1", len(g.Properties))
	}
	if g.TotalProperties != 1 {
		t.Errorf("TotalProperties = %d, want 1", g.TotalProperties)
	}
}

func TestGenerateValidatePhase(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseValidate
	s.LastTestResult = "fail"

	g := Generate(s)

	if g.NextPhase != types.PhaseShrink {
		t.Errorf("NextPhase = %q, want %q (test failed)", g.NextPhase, types.PhaseShrink)
	}
	if g.ExpectedTestResult != "any" {
		t.Errorf("ExpectedTestResult = %q, want %q", g.ExpectedTestResult, "any")
	}
}

func TestGenerateValidatePassPath(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseValidate
	s.LastTestResult = "pass"

	g := Generate(s)

	if g.NextPhase != types.PhaseRefine {
		t.Errorf("NextPhase = %q, want %q (test passed)", g.NextPhase, types.PhaseRefine)
	}
}

func TestGenerateShrinkPhase(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseShrink
	s.ShrinkAnalysis = "found minimal input"

	g := Generate(s)

	if g.ShrinkAnalysis != "found minimal input" {
		t.Errorf("ShrinkAnalysis = %q", g.ShrinkAnalysis)
	}
}

func TestGenerateRefinePhase(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseRefine
	s.Reflections = []types.ReflectionQuestion{
		{ID: 1, Question: "Q1"},
	}
	s.ShrinkAnalysis = "previous analysis"

	g := Generate(s)

	if len(g.Reflections) != 1 {
		t.Errorf("Reflections = %d, want 1", len(g.Reflections))
	}
	if g.ShrinkAnalysis != "previous analysis" {
		t.Errorf("ShrinkAnalysis = %q", g.ShrinkAnalysis)
	}
}

func TestGenerateDonePhase(t *testing.T) {
	s := types.NewSession()
	s.Phase = types.PhaseDone

	g := Generate(s)

	if g.ExpectedTestResult != "" {
		t.Errorf("ExpectedTestResult = %q, want empty for done", g.ExpectedTestResult)
	}
}

func TestGenerateBlockers(t *testing.T) {
	s := types.NewSession()
	// No properties, no selection — should have blockers
	g := Generate(s)

	if len(g.Blockers) == 0 {
		t.Error("Expected blockers for empty session in property phase")
	}
}

func TestGenerateWithTestCmd(t *testing.T) {
	s := types.NewSession()
	s.TestCmd = "go test ./..."

	g := Generate(s)

	if g.TestCmd != "go test ./..." {
		t.Errorf("TestCmd = %q", g.TestCmd)
	}
}
