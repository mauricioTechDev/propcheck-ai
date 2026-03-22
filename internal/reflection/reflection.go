package reflection

import (
	"fmt"
	"strings"

	"github.com/mauricioTechDev/propcheck-ai/internal/types"
)

// MinAnswerWords is the minimum number of words required for a valid answer.
const MinAnswerWords = 5

// DefaultQuestions returns the 7 PBT-specific reflection questions.
func DefaultQuestions() []types.ReflectionQuestion {
	return []types.ReflectionQuestion{
		{ID: 1, Question: "Does my property test an essential invariant, not just restate the implementation?"},
		{ID: 2, Question: "Are my generators producing a representative distribution of inputs, including edge cases?"},
		{ID: 3, Question: "Is my property falsifiable - could a real bug cause it to fail?"},
		{ID: 4, Question: "Have I considered boundary values: empty collections, zero, negative numbers, max values?"},
		{ID: 5, Question: "Can I decompose this property into smaller, independent properties?"},
		{ID: 6, Question: "If a failure was found, have I understood the root cause from the counter-example?"},
		{ID: 7, Question: "Should I add more properties based on what I learned about this function/system?"},
	}
}

// ValidateAnswer checks that an answer has at least MinAnswerWords words.
func ValidateAnswer(answer string) error {
	words := len(strings.Fields(answer))
	if words < MinAnswerWords {
		return fmt.Errorf("answer must be at least %d words, got %d", MinAnswerWords, words)
	}
	return nil
}
