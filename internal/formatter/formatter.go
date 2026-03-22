package formatter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mauricioTechDev/propcheck-ai/internal/phase"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/mauricioTechDev/propcheck-ai/internal/verify"
)

// sortPropertiesByID returns a copy of properties sorted by ID ascending.
func sortPropertiesByID(props []types.Property) []types.Property {
	sorted := make([]types.Property, len(props))
	copy(sorted, props)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

// Format specifies the output format.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// FormatGuidance renders guidance in the specified format.
func FormatGuidance(g types.Guidance, f Format) (string, error) {
	switch f {
	case FormatJSON:
		return formatJSON(g)
	case FormatText:
		return formatText(g), nil
	default:
		return "", fmt.Errorf("unknown format: %q", f)
	}
}

func formatJSON(g types.Guidance) (string, error) {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding guidance: %w", err)
	}
	return string(data), nil
}

func formatText(g types.Guidance) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Phase: %s\n", strings.ToUpper(g.Phase.String()))
	if g.NextPhase != "" {
		fmt.Fprintf(&b, "Next Phase: %s\n", strings.ToUpper(g.NextPhase.String()))
	}
	if g.TestCmd != "" {
		fmt.Fprintf(&b, "Test Command: %s\n", g.TestCmd)
	}
	if g.CurrentProperty != nil {
		cat := ""
		if g.CurrentProperty.Category != "" {
			cat = fmt.Sprintf(" [%s]", g.CurrentProperty.Category)
		}
		fmt.Fprintf(&b, "Current Property: [%d]%s %s\n", g.CurrentProperty.ID, cat, g.CurrentProperty.Description)
	}
	if g.Iteration > 0 {
		fmt.Fprintf(&b, "Iteration: %d\n", g.Iteration)
	}
	if g.ExpectedTestResult != "" {
		fmt.Fprintf(&b, "Expected Test Result: %s\n", g.ExpectedTestResult)
	}
	b.WriteString("\n")

	if len(g.Properties) > 0 {
		b.WriteString("Active Properties:\n")
		for _, p := range sortPropertiesByID(g.Properties) {
			cat := ""
			if p.Category != "" {
				cat = fmt.Sprintf(" [%s]", p.Category)
			}
			fmt.Fprintf(&b, "  [%d]%s %s\n", p.ID, cat, p.Description)
		}
		b.WriteString("\n")
	}

	if len(g.Blockers) > 0 {
		b.WriteString("Blockers:\n")
		for _, bl := range g.Blockers {
			fmt.Fprintf(&b, "  - %s\n", bl)
		}
		b.WriteString("\n")
	}

	if g.ShrinkAnalysis != "" {
		fmt.Fprintf(&b, "Shrink Analysis: %s\n\n", g.ShrinkAnalysis)
	}

	if len(g.Reflections) > 0 {
		answered := 0
		for _, r := range g.Reflections {
			if r.Answer != "" {
				answered++
			}
		}
		fmt.Fprintf(&b, "Reflections (%d/%d answered):\n", answered, len(g.Reflections))
		for _, r := range g.Reflections {
			status := "pending"
			if r.Answer != "" {
				status = "answered"
			}
			fmt.Fprintf(&b, "  [%d] (%s) %s\n", r.ID, status, r.Question)
			if r.Answer != "" {
				fmt.Fprintf(&b, "      -> %q\n", r.Answer)
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// FormatFullStatus renders a rich session overview.
func FormatFullStatus(s *types.Session, f Format) (string, error) {
	type fullStatusOutput struct {
		Phase           types.Phase      `json:"phase"`
		TestCmd         string           `json:"test_cmd,omitempty"`
		CurrentPropID   *int             `json:"current_property_id,omitempty"`
		Iteration       int              `json:"iteration,omitempty"`
		TotalProperties int              `json:"total_properties"`
		ActiveProps     int              `json:"active_properties"`
		DoneProps       int              `json:"done_properties"`
		ComplianceScore *float64         `json:"compliance_score,omitempty"`
		Properties      []types.Property `json:"properties"`
		History         []types.Event    `json:"history,omitempty"`
	}

	active := s.ActiveProperties()
	doneProps := len(s.Properties) - len(active)

	var complianceScore *float64
	if doneProps > 0 {
		result := verify.Analyze(s)
		complianceScore = &result.Score
	}

	out := fullStatusOutput{
		Phase:           s.Phase,
		TestCmd:         s.TestCmd,
		CurrentPropID:   s.CurrentPropertyID,
		Iteration:       s.Iteration,
		TotalProperties: len(s.Properties),
		ActiveProps:     len(active),
		DoneProps:       doneProps,
		ComplianceScore: complianceScore,
		Properties:      s.Properties,
		History:         s.History,
	}

	switch f {
	case FormatJSON:
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case FormatText:
		var b strings.Builder
		fmt.Fprintf(&b, "Phase: %s\n", strings.ToUpper(string(s.Phase)))
		if s.TestCmd != "" {
			fmt.Fprintf(&b, "Test Command: %s\n", s.TestCmd)
		}
		if cs := s.CurrentProperty(); cs != nil {
			fmt.Fprintf(&b, "Current Property: [%d] %s\n", cs.ID, cs.Description)
		}
		if s.Iteration > 0 {
			fmt.Fprintf(&b, "Iteration: %d\n", s.Iteration)
		}
		fmt.Fprintf(&b, "Properties: %d total, %d active, %d done\n", out.TotalProperties, out.ActiveProps, out.DoneProps)
		if complianceScore != nil {
			fmt.Fprintf(&b, "Compliance: %.0f%%\n", *complianceScore)
		}
		b.WriteString("\n")
		for _, prop := range sortPropertiesByID(s.Properties) {
			status := "active"
			if prop.Status == types.PropertyStatusCompleted {
				status = "done"
			}
			cat := ""
			if prop.Category != "" {
				cat = fmt.Sprintf(" [%s]", prop.Category)
			}
			fmt.Fprintf(&b, "  [%d] (%s)%s %s\n", prop.ID, status, cat, prop.Description)
		}
		if len(s.Properties) > 0 {
			b.WriteString("\n")
		}
		if len(s.History) > 0 {
			b.WriteString("History:\n")
			for _, ev := range s.History {
				line := fmt.Sprintf("  %s: %s", ev.Timestamp, ev.Action)
				if ev.From != "" && ev.To != "" {
					line += fmt.Sprintf(" (%s -> %s)", ev.From, ev.To)
				}
				if ev.Result != "" {
					line += fmt.Sprintf(" [%s]", ev.Result)
				}
				if ev.PropCount > 0 {
					line += fmt.Sprintf(" (%d properties)", ev.PropCount)
				}
				fmt.Fprintln(&b, line)
			}
			b.WriteString("\n")
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unknown format: %q", f)
	}
}

// resumeNextAction returns the single most important next action for context recovery.
func resumeNextAction(s *types.Session) string {
	if len(s.Properties) == 0 {
		return `propcheck-ai property add "description"`
	}
	switch s.Phase {
	case types.PhaseDone:
		if len(s.ActiveProperties()) > 0 {
			return "propcheck-ai property done --all"
		}
		return `All properties complete. Add more: propcheck-ai property add "desc"`
	case types.PhaseProperty:
		if s.CurrentPropertyID == nil {
			active := s.ActiveProperties()
			if len(active) > 0 {
				return fmt.Sprintf("propcheck-ai property pick %d", active[0].ID)
			}
		}
		return "propcheck-ai phase next"
	case types.PhaseGenerate:
		return "propcheck-ai phase next"
	case types.PhaseValidate:
		return "propcheck-ai test && propcheck-ai phase next"
	case types.PhaseShrink:
		return `propcheck-ai shrink analyze --answer "your analysis here"`
	case types.PhaseRefine:
		pending := s.PendingReflections()
		if len(pending) > 0 {
			return fmt.Sprintf(`propcheck-ai refine reflect %d --answer "your answer"`, pending[0].ID)
		}
		return "propcheck-ai test && propcheck-ai phase next"
	}
	return "propcheck-ai guide"
}

// recentHistory returns the last n events from the session history.
func recentHistory(s *types.Session, n int) []types.Event {
	if len(s.History) <= n {
		return s.History
	}
	return s.History[len(s.History)-n:]
}

// FormatResume renders a compact session checkpoint for agent context recovery.
func FormatResume(s *types.Session, f Format) (string, error) {
	type resumeOutput struct {
		Phase               types.Phase     `json:"phase"`
		TestCmd             string          `json:"test_cmd,omitempty"`
		Iteration           int             `json:"iteration,omitempty"`
		CurrentProperty     *types.Property `json:"current_property,omitempty"`
		RemainingProperties int             `json:"remaining_properties"`
		Blockers            []string        `json:"blockers,omitempty"`
		NextAction          string          `json:"next_action"`
		RecentEvents        []types.Event   `json:"recent_events,omitempty"`
	}

	remaining := len(s.RemainingProperties())
	blockers := phase.GetBlockers(s)
	recent := recentHistory(s, 5)

	out := resumeOutput{
		Phase:               s.Phase,
		TestCmd:             s.TestCmd,
		Iteration:           s.Iteration,
		RemainingProperties: remaining,
		Blockers:            blockers,
		NextAction:          resumeNextAction(s),
		RecentEvents:        recent,
	}
	if cp := s.CurrentProperty(); cp != nil {
		out.CurrentProperty = cp
	}

	switch f {
	case FormatJSON:
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case FormatText:
		var b strings.Builder
		b.WriteString("=== PBT Session Checkpoint ===\n")
		fmt.Fprintf(&b, "Phase: %s", strings.ToUpper(string(s.Phase)))
		if s.Iteration > 0 {
			fmt.Fprintf(&b, " | Iteration: %d", s.Iteration)
		}
		b.WriteString("\n")
		if cp := s.CurrentProperty(); cp != nil {
			fmt.Fprintf(&b, "Working on: [%d] %s\n", cp.ID, cp.Description)
		} else if s.Phase == types.PhaseProperty && len(s.ActiveProperties()) > 0 {
			b.WriteString("Working on: (no property selected)\n")
		}
		if remaining > 0 {
			fmt.Fprintf(&b, "Remaining properties: %d\n", remaining)
		}
		b.WriteString("\n")
		if len(blockers) > 0 {
			b.WriteString("BLOCKERS:\n")
			for _, bl := range blockers {
				fmt.Fprintf(&b, "  - %s\n", bl)
			}
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "NEXT ACTION:\n  %s\n", out.NextAction)
		if len(recent) > 0 {
			b.WriteString("\nRecent events:\n")
			for _, ev := range recent {
				line := "  " + ev.Action
				if ev.From != "" && ev.To != "" {
					line += fmt.Sprintf(" (%s -> %s)", ev.From, ev.To)
				}
				if ev.Result != "" {
					line += fmt.Sprintf(" [%s]", ev.Result)
				}
				fmt.Fprintln(&b, line)
			}
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unknown format: %q", f)
	}
}

// FormatStatus renders a simple session status.
func FormatStatus(s *types.Session, f Format) (string, error) {
	type statusOutput struct {
		Phase           types.Phase      `json:"phase"`
		TotalProperties int              `json:"total_properties"`
		ActiveProps     int              `json:"active_properties"`
		DoneProps       int              `json:"done_properties"`
		Properties      []types.Property `json:"properties"`
	}

	active := s.ActiveProperties()
	out := statusOutput{
		Phase:           s.Phase,
		TotalProperties: len(s.Properties),
		ActiveProps:     len(active),
		DoneProps:       len(s.Properties) - len(active),
		Properties:      s.Properties,
	}

	switch f {
	case FormatJSON:
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case FormatText:
		var b strings.Builder
		fmt.Fprintf(&b, "Phase: %s\n", strings.ToUpper(string(s.Phase)))
		fmt.Fprintf(&b, "Properties: %d total, %d active, %d done\n\n", out.TotalProperties, out.ActiveProps, out.DoneProps)
		for _, prop := range sortPropertiesByID(s.Properties) {
			status := "active"
			if prop.Status == types.PropertyStatusCompleted {
				status = "done"
			}
			cat := ""
			if prop.Category != "" {
				cat = fmt.Sprintf(" [%s]", prop.Category)
			}
			isCurrent := s.CurrentPropertyID != nil && prop.ID == *s.CurrentPropertyID
			if isCurrent {
				fmt.Fprintf(&b, "\u2192 [%d] (%s)%s %s (current)\n", prop.ID, status, cat, prop.Description)
			} else {
				fmt.Fprintf(&b, "  [%d] (%s)%s %s\n", prop.ID, status, cat, prop.Description)
			}
		}
		if len(s.Properties) > 0 {
			b.WriteString("\n")
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unknown format: %q", f)
	}
}
