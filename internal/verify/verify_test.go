package verify

import (
	"testing"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

func compliantSession() *types.Session {
	s := types.NewSession()
	s.AddProperty("sort idempotent", "invariant")
	s.SetCurrentProperty(1)

	s.AddEvent("property_picked", func(e *types.Event) { e.PropertyID = 1 })
	s.AddEvent("phase_next", func(e *types.Event) { e.From = "property"; e.To = "generate" })
	s.AddEvent("phase_next", func(e *types.Event) { e.From = "generate"; e.To = "validate" })
	s.AddEvent("test_run", func(e *types.Event) { e.Result = "pass" })
	s.AddEvent("phase_next", func(e *types.Event) { e.From = "validate"; e.To = "refine" })
	s.AddEvent("phase_next", func(e *types.Event) { e.From = "refine"; e.To = "done" })

	s.CompleteProperty(1)
	return s
}

func TestAnalyzeCompliant(t *testing.T) {
	s := compliantSession()
	r := Analyze(s)

	if !r.Compliant {
		t.Errorf("expected compliant, got violations: %v", r.Violations)
	}
	if r.Score != 100 {
		t.Errorf("Score = %f, want 100", r.Score)
	}
	if r.PropertiesVerified != 1 {
		t.Errorf("PropertiesVerified = %d, want 1", r.PropertiesVerified)
	}
}

func TestAnalyzeMissingPicked(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("test", "")
	s.CompleteProperty(1)
	// No property_picked event

	r := Analyze(s)
	if r.Compliant {
		t.Error("expected non-compliant")
	}
	found := false
	for _, v := range r.Violations {
		if v.Rule == "property_picked_required" {
			found = true
		}
	}
	if !found {
		t.Error("expected property_picked_required violation")
	}
}

func TestAnalyzeMissingTestRun(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("test", "")
	s.AddEvent("property_picked", func(e *types.Event) { e.PropertyID = 1 })
	// No test_run event
	s.CompleteProperty(1)

	r := Analyze(s)
	found := false
	for _, v := range r.Violations {
		if v.Rule == "validate_test_required" {
			found = true
		}
	}
	if !found {
		t.Error("expected validate_test_required violation")
	}
}

func TestAnalyzeMissingShrinkAnalysis(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("test", "")
	s.AddEvent("property_picked", func(e *types.Event) { e.PropertyID = 1 })
	s.AddEvent("test_run", func(e *types.Event) { e.Result = "fail" })
	// No shrink_analyze event
	s.CompleteProperty(1)

	r := Analyze(s)
	found := false
	for _, v := range r.Violations {
		if v.Rule == "shrink_analysis_required" {
			found = true
		}
	}
	if !found {
		t.Error("expected shrink_analysis_required violation")
	}
}

func TestAnalyzeWithShrinkAnalysis(t *testing.T) {
	s := types.NewSession()
	s.AddProperty("test", "")
	s.AddEvent("property_picked", func(e *types.Event) { e.PropertyID = 1 })
	s.AddEvent("test_run", func(e *types.Event) { e.Result = "fail" })
	s.AddEvent("shrink_analyze", func(e *types.Event) { e.Result = "recorded" })
	s.AddEvent("test_run", func(e *types.Event) { e.Result = "pass" })
	s.CompleteProperty(1)

	r := Analyze(s)
	// Should not have shrink_analysis_required violation
	for _, v := range r.Violations {
		if v.Rule == "shrink_analysis_required" {
			t.Error("should not have shrink_analysis_required when analysis exists")
		}
	}
}

func TestAnalyzePhaseSetViolation(t *testing.T) {
	s := compliantSession()
	s.AddEvent("phase_set", func(e *types.Event) {
		e.From = "property"
		e.To = "validate"
		e.Result = "forced_override"
	})

	r := Analyze(s)
	if r.Compliant {
		t.Error("expected non-compliant with phase_set")
	}
	found := false
	for _, v := range r.Violations {
		if v.Rule == "no_phase_set" {
			found = true
		}
	}
	if !found {
		t.Error("expected no_phase_set violation")
	}
}

func TestAnalyzeMixedCompliance(t *testing.T) {
	s := types.NewSession()

	// Property 1: compliant
	s.AddProperty("a", "")
	s.AddEvent("property_picked", func(e *types.Event) { e.PropertyID = 1 })
	s.AddEvent("test_run", func(e *types.Event) { e.Result = "pass" })
	s.CompleteProperty(1)

	// Property 2: no picked event
	s.AddProperty("b", "")
	s.CompleteProperty(2)

	r := Analyze(s)
	if r.PropertiesVerified != 2 {
		t.Errorf("PropertiesVerified = %d, want 2", r.PropertiesVerified)
	}
	if r.PropertiesCompliant != 1 {
		t.Errorf("PropertiesCompliant = %d, want 1", r.PropertiesCompliant)
	}
	if r.Score != 50 {
		t.Errorf("Score = %f, want 50", r.Score)
	}
}

func TestAnalyzeNoProperties(t *testing.T) {
	s := types.NewSession()
	r := Analyze(s)

	if !r.Compliant {
		t.Error("empty session should be compliant")
	}
	if r.Score != 100 {
		t.Errorf("Score = %f, want 100", r.Score)
	}
}
