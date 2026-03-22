package reflection

import (
	"testing"
)

func TestDefaultQuestions(t *testing.T) {
	qs := DefaultQuestions()
	if len(qs) != 7 {
		t.Errorf("DefaultQuestions() returned %d questions, want 7", len(qs))
	}
	for i, q := range qs {
		if q.ID != i+1 {
			t.Errorf("Question %d has ID %d, want %d", i, q.ID, i+1)
		}
		if q.Question == "" {
			t.Errorf("Question %d has empty text", q.ID)
		}
		if q.Answer != "" {
			t.Errorf("Question %d has non-empty answer", q.ID)
		}
	}
}

func TestValidateAnswer(t *testing.T) {
	tests := []struct {
		answer string
		valid  bool
	}{
		{"", false},
		{"yes", false},
		{"two words", false},
		{"three words here", false},
		{"four words are here", false},
		{"five words are right here", true},
		{"this answer has more than five words in it", true},
	}
	for _, tt := range tests {
		err := ValidateAnswer(tt.answer)
		if tt.valid && err != nil {
			t.Errorf("ValidateAnswer(%q) = %v, want nil", tt.answer, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateAnswer(%q) = nil, want error", tt.answer)
		}
	}
}
