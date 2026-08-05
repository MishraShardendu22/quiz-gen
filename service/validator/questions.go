package validator

import (
	"fmt"
	"strings"

	"github.com/MishraShardendu22/quiz-gen/model"
)

// ValidateQuestion validates a single question against all requirements
func ValidateQuestion(q model.LLMQuestion) error {
	// Check question is not empty
	if strings.TrimSpace(q.Question) == "" {
		return fmt.Errorf("question text is empty")
	}

	// Check exactly 4 options
	if len(q.Options) != 4 {
		return fmt.Errorf("expected 4 options, got %d", len(q.Options))
	}

	// Check no option is empty
	for i, opt := range q.Options {
		if strings.TrimSpace(opt) == "" {
			return fmt.Errorf("option %d is empty", i+1)
		}
	}

	// Check correct answer is valid (0-3)
	if q.CorrectAnswer < 0 || q.CorrectAnswer > 3 {
		return fmt.Errorf("correct_answer must be 0-3, got %d", q.CorrectAnswer)
	}

	// Check explanation is not empty
	if strings.TrimSpace(q.Explanation) == "" {
		return fmt.Errorf("explanation is empty")
	}

	return nil
}

// ValidateQuestions validates a slice of questions
func ValidateQuestions(questions []model.LLMQuestion) error {
	if len(questions) == 0 {
		return fmt.Errorf("no questions to validate")
	}

	for i, q := range questions {
		if err := ValidateQuestion(q); err != nil {
			return fmt.Errorf("question %d validation failed: %w", i+1, err)
		}
	}

	return nil
}
