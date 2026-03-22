package phase

import (
	"testing"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

func TestNext(t *testing.T) {
	tests := []struct {
		from types.Phase
		to   types.Phase
	}{
		{types.PhaseProperty, types.PhaseGenerate},
		{types.PhaseGenerate, types.PhaseValidate},
		{types.PhaseValidate, types.PhaseRefine},
		{types.PhaseRefine, types.PhaseDone},
	}
	for _, tt := range tests {
		t.Run(string(tt.from), func(t *testing.T) {
			got, err := Next(tt.from)
			if err != nil {
				t.Fatalf("Next(%q) error: %v", tt.from, err)
			}
			if got != tt.to {
				t.Errorf("Next(%q) = %q, want %q", tt.from, got, tt.to)
			}
		})
	}
}

func TestNextFromDoneErrors(t *testing.T) {
	_, err := Next(types.PhaseDone)
	if err == nil {
		t.Error("Next(done) should error")
	}
}

func TestNextWithResultValidatePass(t *testing.T) {
	got, err := NextWithResult(types.PhaseValidate, "pass")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseRefine {
		t.Errorf("got %q, want %q", got, types.PhaseRefine)
	}
}

func TestNextWithResultValidateFail(t *testing.T) {
	got, err := NextWithResult(types.PhaseValidate, "fail")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseShrink {
		t.Errorf("got %q, want %q", got, types.PhaseShrink)
	}
}

func TestNextWithResultShrink(t *testing.T) {
	got, err := NextWithResult(types.PhaseShrink, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseRefine {
		t.Errorf("got %q, want %q", got, types.PhaseRefine)
	}
}

func TestNextWithResultNonValidateIgnoresResult(t *testing.T) {
	got, err := NextWithResult(types.PhaseProperty, "fail")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseGenerate {
		t.Errorf("got %q, want %q", got, types.PhaseGenerate)
	}
}

func TestNextInLoopFromRefineWithRemaining(t *testing.T) {
	got, err := NextInLoop(types.PhaseRefine, "pass", true)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseProperty {
		t.Errorf("got %q, want %q", got, types.PhaseProperty)
	}
}

func TestNextInLoopFromRefineNoRemaining(t *testing.T) {
	got, err := NextInLoop(types.PhaseRefine, "pass", false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseDone {
		t.Errorf("got %q, want %q", got, types.PhaseDone)
	}
}

func TestNextInLoopFromValidateFail(t *testing.T) {
	got, err := NextInLoop(types.PhaseValidate, "fail", true)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseShrink {
		t.Errorf("got %q, want %q", got, types.PhaseShrink)
	}
}

func TestNextInLoopFromValidatePass(t *testing.T) {
	got, err := NextInLoop(types.PhaseValidate, "pass", true)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != types.PhaseRefine {
		t.Errorf("got %q, want %q", got, types.PhaseRefine)
	}
}

func TestExpectedTestResult(t *testing.T) {
	tests := []struct {
		phase types.Phase
		want  string
	}{
		{types.PhaseProperty, ""},
		{types.PhaseGenerate, ""},
		{types.PhaseValidate, "any"},
		{types.PhaseShrink, "fail"},
		{types.PhaseRefine, "pass"},
		{types.PhaseDone, ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			got := ExpectedTestResult(tt.phase)
			if got != tt.want {
				t.Errorf("ExpectedTestResult(%q) = %q, want %q", tt.phase, got, tt.want)
			}
		})
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from  types.Phase
		to    types.Phase
		valid bool
	}{
		{types.PhaseProperty, types.PhaseGenerate, true},
		{types.PhaseGenerate, types.PhaseValidate, true},
		{types.PhaseValidate, types.PhaseRefine, true},
		{types.PhaseValidate, types.PhaseShrink, true},
		{types.PhaseShrink, types.PhaseRefine, true},
		{types.PhaseRefine, types.PhaseDone, true},
		{types.PhaseRefine, types.PhaseProperty, true},
		// Invalid
		{types.PhaseProperty, types.PhaseValidate, false},
		{types.PhaseGenerate, types.PhaseRefine, false},
		{types.PhaseDone, types.PhaseProperty, false},
		{types.PhaseShrink, types.PhaseProperty, false},
	}
	for _, tt := range tests {
		name := string(tt.from) + "->" + string(tt.to)
		t.Run(name, func(t *testing.T) {
			got := CanTransition(tt.from, tt.to)
			if got != tt.valid {
				t.Errorf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.valid)
			}
		})
	}
}
