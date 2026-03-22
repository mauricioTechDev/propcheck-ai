package types

import (
	"fmt"
	"time"
)

// Phase represents a stage in the PBT cycle.
type Phase string

const (
	PhaseProperty Phase = "property"
	PhaseGenerate Phase = "generate"
	PhaseValidate Phase = "validate"
	PhaseShrink   Phase = "shrink"
	PhaseRefine   Phase = "refine"
	PhaseDone     Phase = "done"
)

// ValidPhases returns all valid phase values.
func ValidPhases() []Phase {
	return []Phase{PhaseProperty, PhaseGenerate, PhaseValidate, PhaseShrink, PhaseRefine, PhaseDone}
}

// IsValid checks whether the phase is a recognized value.
func (p Phase) IsValid() bool {
	for _, v := range ValidPhases() {
		if p == v {
			return true
		}
	}
	return false
}

// String returns the string representation of a Phase.
func (p Phase) String() string {
	return string(p)
}

// PropertyStatus represents the state of a property.
type PropertyStatus string

const (
	PropertyStatusActive    PropertyStatus = "active"
	PropertyStatusCompleted PropertyStatus = "completed"
)

// Property is a single property-based test to be implemented.
type Property struct {
	ID          int            `json:"id"`
	Description string         `json:"description"`
	Status      PropertyStatus `json:"status"`
	Category    string         `json:"category,omitempty"`
}

// ValidCategories returns the allowed property categories.
func ValidCategories() []string {
	return []string{"invariant", "roundtrip", "equivalence", "metamorphic"}
}

// ReflectionQuestion is a structured prompt the agent must answer during the refine phase.
type ReflectionQuestion struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer,omitempty"`
}

// Event records a notable action during the PBT session for audit trail.
type Event struct {
	Action     string `json:"action"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Result     string `json:"result,omitempty"`
	PropCount  int    `json:"prop_count,omitempty"`
	PropertyID int    `json:"property_id,omitempty"`
	Timestamp  string `json:"at"`
}

// Session holds the full state of a PBT session.
type Session struct {
	Phase             Phase                `json:"phase"`
	AgentMode         bool                 `json:"agent_mode,omitempty"`
	TestCmd           string               `json:"test_cmd,omitempty"`
	LastTestResult    string               `json:"last_test_result,omitempty"`
	Properties        []Property           `json:"properties"`
	NextID            int                  `json:"next_id"`
	CurrentPropertyID *int                 `json:"current_property_id,omitempty"`
	Iteration         int                  `json:"iteration,omitempty"`
	ShrinkAnalysis    string               `json:"shrink_analysis,omitempty"`
	Reflections       []ReflectionQuestion `json:"reflections,omitempty"`
	History           []Event              `json:"history,omitempty"`
}

// NewSession creates a fresh PBT session starting in the property phase.
func NewSession() *Session {
	return &Session{
		Phase:      PhaseProperty,
		Properties: []Property{},
		NextID:     1,
	}
}

// AddProperty adds a new property to the session and returns the assigned ID.
func (s *Session) AddProperty(description string, category string) int {
	id := s.NextID
	p := Property{
		ID:          id,
		Description: description,
		Status:      PropertyStatusActive,
	}
	if category != "" {
		p.Category = category
	}
	s.Properties = append(s.Properties, p)
	s.NextID++
	return id
}

// CompleteProperty marks a property as completed by ID.
func (s *Session) CompleteProperty(id int) error {
	for i, p := range s.Properties {
		if p.ID == id {
			if p.Status == PropertyStatusCompleted {
				return fmt.Errorf("property %d is already completed", id)
			}
			s.Properties[i].Status = PropertyStatusCompleted
			return nil
		}
	}
	return fmt.Errorf("property %d not found", id)
}

// CompleteAllProperties marks all active properties as completed and returns the count.
func (s *Session) CompleteAllProperties() int {
	count := 0
	for i, p := range s.Properties {
		if p.Status == PropertyStatusActive {
			s.Properties[i].Status = PropertyStatusCompleted
			count++
		}
	}
	return count
}

// ActiveProperties returns only properties that are not yet completed.
func (s *Session) ActiveProperties() []Property {
	var active []Property
	for _, p := range s.Properties {
		if p.Status == PropertyStatusActive {
			active = append(active, p)
		}
	}
	return active
}

// CurrentProperty returns the property matching CurrentPropertyID, or nil if none is set.
func (s *Session) CurrentProperty() *Property {
	if s.CurrentPropertyID == nil {
		return nil
	}
	for i, p := range s.Properties {
		if p.ID == *s.CurrentPropertyID {
			return &s.Properties[i]
		}
	}
	return nil
}

// SetCurrentProperty sets CurrentPropertyID after validating the property exists and is active.
func (s *Session) SetCurrentProperty(id int) error {
	for _, p := range s.Properties {
		if p.ID == id {
			if p.Status != PropertyStatusActive {
				return fmt.Errorf("property %d is not active", id)
			}
			s.CurrentPropertyID = &id
			return nil
		}
	}
	return fmt.Errorf("property %d not found", id)
}

// CompleteCurrentProperty marks the current property as completed and clears CurrentPropertyID.
func (s *Session) CompleteCurrentProperty() error {
	if s.CurrentPropertyID == nil {
		return fmt.Errorf("no current property selected")
	}
	if err := s.CompleteProperty(*s.CurrentPropertyID); err != nil {
		return err
	}
	s.CurrentPropertyID = nil
	return nil
}

// RemainingProperties returns active properties excluding the current one.
func (s *Session) RemainingProperties() []Property {
	var remaining []Property
	for _, p := range s.Properties {
		if p.Status == PropertyStatusActive && (s.CurrentPropertyID == nil || p.ID != *s.CurrentPropertyID) {
			remaining = append(remaining, p)
		}
	}
	return remaining
}

// PendingReflections returns reflection questions that have not been answered.
func (s *Session) PendingReflections() []ReflectionQuestion {
	var pending []ReflectionQuestion
	for _, r := range s.Reflections {
		if r.Answer == "" {
			pending = append(pending, r)
		}
	}
	return pending
}

// AllReflectionsAnswered returns true when all reflection questions have answers,
// or when the reflections slice is empty.
func (s *Session) AllReflectionsAnswered() bool {
	for _, r := range s.Reflections {
		if r.Answer == "" {
			return false
		}
	}
	return true
}

// AnswerReflection sets the answer for a reflection question by ID.
func (s *Session) AnswerReflection(id int, answer string) error {
	for i, r := range s.Reflections {
		if r.ID == id {
			s.Reflections[i].Answer = answer
			return nil
		}
	}
	return fmt.Errorf("reflection question %d not found", id)
}

// AddEvent appends an event to the session history.
func (s *Session) AddEvent(action string, opts ...func(*Event)) {
	e := Event{
		Action:    action,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	for _, opt := range opts {
		opt(&e)
	}
	s.History = append(s.History, e)
}

// Guidance is the structured output of the guide command.
type Guidance struct {
	Phase              Phase                `json:"phase"`
	NextPhase          Phase                `json:"next_phase,omitempty"`
	TestCmd            string               `json:"test_cmd,omitempty"`
	Properties         []Property           `json:"properties"`
	CurrentProperty    *Property            `json:"current_property,omitempty"`
	Iteration          int                  `json:"iteration,omitempty"`
	TotalProperties    int                  `json:"total_properties,omitempty"`
	ExpectedTestResult string               `json:"expected_test_result,omitempty"`
	Blockers           []string             `json:"blockers,omitempty"`
	Reflections        []ReflectionQuestion `json:"reflections,omitempty"`
	ShrinkAnalysis     string               `json:"shrink_analysis,omitempty"`
}
