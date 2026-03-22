package types

import (
	"encoding/json"
	"testing"
)

func TestPhaseIsValid(t *testing.T) {
	tests := []struct {
		phase Phase
		valid bool
	}{
		{PhaseProperty, true},
		{PhaseGenerate, true},
		{PhaseValidate, true},
		{PhaseShrink, true},
		{PhaseRefine, true},
		{PhaseDone, true},
		{Phase("invalid"), false},
		{Phase(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := tt.phase.IsValid(); got != tt.valid {
				t.Errorf("Phase(%q).IsValid() = %v, want %v", tt.phase, got, tt.valid)
			}
		})
	}
}

func TestPhaseString(t *testing.T) {
	if got := PhaseProperty.String(); got != "property" {
		t.Errorf("PhaseProperty.String() = %q, want %q", got, "property")
	}
}

func TestValidPhases(t *testing.T) {
	phases := ValidPhases()
	if len(phases) != 6 {
		t.Errorf("ValidPhases() returned %d phases, want 6", len(phases))
	}
}

func TestNewSession(t *testing.T) {
	s := NewSession()
	if s.Phase != PhaseProperty {
		t.Errorf("NewSession().Phase = %q, want %q", s.Phase, PhaseProperty)
	}
	if s.NextID != 1 {
		t.Errorf("NewSession().NextID = %d, want 1", s.NextID)
	}
	if len(s.Properties) != 0 {
		t.Errorf("NewSession().Properties has %d items, want 0", len(s.Properties))
	}
	if s.CurrentPropertyID != nil {
		t.Error("NewSession().CurrentPropertyID should be nil")
	}
}

func TestAddProperty(t *testing.T) {
	s := NewSession()
	id1 := s.AddProperty("sort is idempotent", "invariant")
	id2 := s.AddProperty("reverse roundtrip", "roundtrip")
	id3 := s.AddProperty("no category", "")

	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Errorf("IDs = %d, %d, %d; want 1, 2, 3", id1, id2, id3)
	}
	if len(s.Properties) != 3 {
		t.Fatalf("Properties count = %d, want 3", len(s.Properties))
	}
	if s.Properties[0].Category != "invariant" {
		t.Errorf("Properties[0].Category = %q, want %q", s.Properties[0].Category, "invariant")
	}
	if s.Properties[2].Category != "" {
		t.Errorf("Properties[2].Category = %q, want empty", s.Properties[2].Category)
	}
	if s.Properties[0].Status != PropertyStatusActive {
		t.Errorf("Properties[0].Status = %q, want %q", s.Properties[0].Status, PropertyStatusActive)
	}
}

func TestCompleteProperty(t *testing.T) {
	s := NewSession()
	s.AddProperty("test", "")

	if err := s.CompleteProperty(1); err != nil {
		t.Fatalf("CompleteProperty(1) = %v", err)
	}
	if s.Properties[0].Status != PropertyStatusCompleted {
		t.Errorf("Status = %q, want %q", s.Properties[0].Status, PropertyStatusCompleted)
	}

	// Already completed
	if err := s.CompleteProperty(1); err == nil {
		t.Error("CompleteProperty(1) should error on already completed")
	}

	// Not found
	if err := s.CompleteProperty(99); err == nil {
		t.Error("CompleteProperty(99) should error on not found")
	}
}

func TestCompleteAllProperties(t *testing.T) {
	s := NewSession()
	s.AddProperty("a", "")
	s.AddProperty("b", "")
	s.AddProperty("c", "")
	s.CompleteProperty(1)

	count := s.CompleteAllProperties()
	if count != 2 {
		t.Errorf("CompleteAllProperties() = %d, want 2", count)
	}
}

func TestActiveProperties(t *testing.T) {
	s := NewSession()
	s.AddProperty("a", "")
	s.AddProperty("b", "")
	s.CompleteProperty(1)

	active := s.ActiveProperties()
	if len(active) != 1 {
		t.Fatalf("ActiveProperties() = %d items, want 1", len(active))
	}
	if active[0].ID != 2 {
		t.Errorf("ActiveProperties()[0].ID = %d, want 2", active[0].ID)
	}
}

func TestSetAndGetCurrentProperty(t *testing.T) {
	s := NewSession()
	s.AddProperty("test", "")

	if err := s.SetCurrentProperty(1); err != nil {
		t.Fatalf("SetCurrentProperty(1) = %v", err)
	}
	cp := s.CurrentProperty()
	if cp == nil {
		t.Fatal("CurrentProperty() = nil")
	}
	if cp.ID != 1 {
		t.Errorf("CurrentProperty().ID = %d, want 1", cp.ID)
	}

	// Not found
	if err := s.SetCurrentProperty(99); err == nil {
		t.Error("SetCurrentProperty(99) should error")
	}

	// Not active
	s.CompleteProperty(1)
	if err := s.SetCurrentProperty(1); err == nil {
		t.Error("SetCurrentProperty on completed should error")
	}
}

func TestCurrentPropertyNil(t *testing.T) {
	s := NewSession()
	if cp := s.CurrentProperty(); cp != nil {
		t.Error("CurrentProperty() should be nil when no property selected")
	}
}

func TestCompleteCurrentProperty(t *testing.T) {
	s := NewSession()
	s.AddProperty("test", "")
	s.SetCurrentProperty(1)

	if err := s.CompleteCurrentProperty(); err != nil {
		t.Fatalf("CompleteCurrentProperty() = %v", err)
	}
	if s.CurrentPropertyID != nil {
		t.Error("CurrentPropertyID should be nil after CompleteCurrentProperty")
	}
	if s.Properties[0].Status != PropertyStatusCompleted {
		t.Errorf("Status = %q, want completed", s.Properties[0].Status)
	}

	// No current selected
	if err := s.CompleteCurrentProperty(); err == nil {
		t.Error("CompleteCurrentProperty() should error when no current")
	}
}

func TestRemainingProperties(t *testing.T) {
	s := NewSession()
	s.AddProperty("a", "")
	s.AddProperty("b", "")
	s.AddProperty("c", "")
	s.SetCurrentProperty(1)

	remaining := s.RemainingProperties()
	if len(remaining) != 2 {
		t.Fatalf("RemainingProperties() = %d, want 2", len(remaining))
	}
	for _, r := range remaining {
		if r.ID == 1 {
			t.Error("RemainingProperties should not include current property")
		}
	}
}

func TestReflections(t *testing.T) {
	s := NewSession()
	s.Reflections = []ReflectionQuestion{
		{ID: 1, Question: "Q1"},
		{ID: 2, Question: "Q2"},
	}

	if s.AllReflectionsAnswered() {
		t.Error("AllReflectionsAnswered() should be false with unanswered questions")
	}

	pending := s.PendingReflections()
	if len(pending) != 2 {
		t.Errorf("PendingReflections() = %d, want 2", len(pending))
	}

	s.AnswerReflection(1, "This is my answer here")
	pending = s.PendingReflections()
	if len(pending) != 1 {
		t.Errorf("PendingReflections() after 1 answer = %d, want 1", len(pending))
	}

	s.AnswerReflection(2, "Another good answer here")
	if !s.AllReflectionsAnswered() {
		t.Error("AllReflectionsAnswered() should be true after all answered")
	}

	// Not found
	if err := s.AnswerReflection(99, "nope"); err == nil {
		t.Error("AnswerReflection(99) should error")
	}
}

func TestReflectionsEmptySlice(t *testing.T) {
	s := NewSession()
	if !s.AllReflectionsAnswered() {
		t.Error("AllReflectionsAnswered() should be true for empty slice")
	}
}

func TestAddEvent(t *testing.T) {
	s := NewSession()
	s.AddEvent("init")
	s.AddEvent("property_add", func(e *Event) {
		e.PropCount = 2
	})
	s.AddEvent("phase_next", func(e *Event) {
		e.From = "property"
		e.To = "generate"
	})

	if len(s.History) != 3 {
		t.Fatalf("History has %d events, want 3", len(s.History))
	}
	if s.History[0].Action != "init" {
		t.Errorf("History[0].Action = %q, want %q", s.History[0].Action, "init")
	}
	if s.History[1].PropCount != 2 {
		t.Errorf("History[1].PropCount = %d, want 2", s.History[1].PropCount)
	}
	if s.History[2].From != "property" || s.History[2].To != "generate" {
		t.Errorf("History[2] transition = %s->%s, want property->generate", s.History[2].From, s.History[2].To)
	}
	if s.History[0].Timestamp == "" {
		t.Error("Event should have a timestamp")
	}
}

func TestSessionJSONRoundTrip(t *testing.T) {
	s := NewSession()
	s.AddProperty("sort idempotent", "invariant")
	s.SetCurrentProperty(1)
	s.ShrinkAnalysis = "counter-example found"
	s.AddEvent("init")

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var s2 Session
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if s2.Phase != PhaseProperty {
		t.Errorf("Phase = %q, want %q", s2.Phase, PhaseProperty)
	}
	if len(s2.Properties) != 1 {
		t.Errorf("Properties = %d, want 1", len(s2.Properties))
	}
	if s2.Properties[0].Category != "invariant" {
		t.Errorf("Category = %q, want %q", s2.Properties[0].Category, "invariant")
	}
	if s2.CurrentPropertyID == nil || *s2.CurrentPropertyID != 1 {
		t.Error("CurrentPropertyID should be 1 after roundtrip")
	}
	if s2.ShrinkAnalysis != "counter-example found" {
		t.Errorf("ShrinkAnalysis = %q", s2.ShrinkAnalysis)
	}
}

func TestSessionOmitsEmptyFields(t *testing.T) {
	s := NewSession()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	// These fields should be omitted when empty
	for _, key := range []string{"agent_mode", "test_cmd", "last_test_result", "current_property_id", "iteration", "shrink_analysis", "reflections", "history"} {
		if _, exists := raw[key]; exists {
			t.Errorf("field %q should be omitted when empty/zero", key)
		}
	}

	// These should always be present
	for _, key := range []string{"phase", "properties", "next_id"} {
		if _, exists := raw[key]; !exists {
			t.Errorf("field %q should always be present", key)
		}
	}
}

func TestValidCategories(t *testing.T) {
	cats := ValidCategories()
	if len(cats) != 4 {
		t.Errorf("ValidCategories() = %d, want 4", len(cats))
	}
}
