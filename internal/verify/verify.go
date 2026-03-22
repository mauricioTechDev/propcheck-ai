package verify

import (
	"fmt"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

// Violation represents a PBT compliance violation.
type Violation struct {
	PropertyID int    `json:"property_id,omitempty"`
	Rule       string `json:"rule"`
	Message    string `json:"message"`
}

// Result holds the compliance analysis output.
type Result struct {
	Violations          []Violation `json:"violations"`
	PropertiesVerified  int         `json:"properties_verified"`
	PropertiesCompliant int         `json:"properties_compliant"`
	Score               float64     `json:"score"`
	Compliant           bool        `json:"compliant"`
}

// Analyze performs post-hoc PBT compliance checking.
func Analyze(s *types.Session) Result {
	var violations []Violation

	// Global: check for phase_set usage
	for _, ev := range s.History {
		if ev.Action == "phase_set" {
			violations = append(violations, Violation{
				Rule:    "no_phase_set",
				Message: fmt.Sprintf("phase_set used (%s -> %s) — bypasses PBT guardrails", ev.From, ev.To),
			})
		}
	}

	// Per-property analysis
	completedIDs := completedPropertyIDs(s)
	propsCompliant := 0

	for _, propID := range completedIDs {
		propViolations := analyzeProperty(s, propID)
		if len(propViolations) == 0 {
			propsCompliant++
		}
		violations = append(violations, propViolations...)
	}

	score := float64(100)
	if len(completedIDs) > 0 {
		score = float64(propsCompliant) / float64(len(completedIDs)) * 100
	}

	if violations == nil {
		violations = []Violation{}
	}

	return Result{
		Violations:          violations,
		PropertiesVerified:  len(completedIDs),
		PropertiesCompliant: propsCompliant,
		Score:               score,
		Compliant:           len(violations) == 0,
	}
}

// completedPropertyIDs returns the IDs of all completed properties.
func completedPropertyIDs(s *types.Session) []int {
	var ids []int
	for _, p := range s.Properties {
		if p.Status == types.PropertyStatusCompleted {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// analyzeProperty checks compliance for a single completed property.
func analyzeProperty(s *types.Session, propID int) []Violation {
	var violations []Violation

	// Check for property_picked event
	hasPicked := false
	hasTestRun := false
	hasFailure := false
	hasShrinkAnalyze := false
	inScope := false

	for _, ev := range s.History {
		if ev.Action == "property_picked" && ev.PropertyID == propID {
			hasPicked = true
			inScope = true
			hasTestRun = false
			hasFailure = false
			hasShrinkAnalyze = false
			continue
		}
		if !inScope {
			continue
		}
		// Stop scope at next property_picked or completion
		if ev.Action == "property_picked" && ev.PropertyID != propID {
			break
		}
		if ev.Action == "test_run" {
			hasTestRun = true
			if ev.Result == "fail" {
				hasFailure = true
			}
		}
		if ev.Action == "shrink_analyze" {
			hasShrinkAnalyze = true
		}
	}

	if !hasPicked {
		violations = append(violations, Violation{
			PropertyID: propID,
			Rule:       "property_picked_required",
			Message:    fmt.Sprintf("property %d completed without property_picked event", propID),
		})
	}

	if hasPicked && !hasTestRun {
		violations = append(violations, Violation{
			PropertyID: propID,
			Rule:       "validate_test_required",
			Message:    fmt.Sprintf("property %d has no test_run event during validation", propID),
		})
	}

	if hasFailure && !hasShrinkAnalyze {
		violations = append(violations, Violation{
			PropertyID: propID,
			Rule:       "shrink_analysis_required",
			Message:    fmt.Sprintf("property %d had a test failure but no shrink analysis recorded", propID),
		})
	}

	return violations
}
