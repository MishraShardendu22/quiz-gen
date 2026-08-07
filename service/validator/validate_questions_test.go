package validator

import (
	"testing"

	"github.com/MishraShardendu22/quiz-gen/model"
)

func TestValidateQuestions_AllValid(t *testing.T) {
	questions := []model.LLMQuestion{
		{
			Question:      "How often should fire extinguishers be inspected?",
			Options:       []string{"Daily", "Weekly", "Monthly", "Yearly"},
			CorrectAnswer: 2,
			Explanation:   "Inspect pressure gauge monthly per safety guidelines.",
		},
		{
			Question:      "What color is a standard water fire extinguisher?",
			Options:       []string{"Red", "Cream", "Blue", "Black"},
			CorrectAnswer: 0,
			Explanation:   "Water extinguishers in standard color coding are signal red.",
		},
	}

	if err := ValidateQuestions(questions); err != nil {
		t.Fatalf("expected all valid questions to pass, got: %v", err)
	}
}

func TestValidateQuestions_OneInvalid(t *testing.T) {
	questions := []model.LLMQuestion{
		{
			Question:      "How often should fire extinguishers be inspected?",
			Options:       []string{"Daily", "Weekly", "Monthly", "Yearly"},
			CorrectAnswer: 2,
			Explanation:   "Inspect pressure gauge monthly per safety guidelines.",
		},
		{
			Question:      "Invalid question with no options",
			Options:       []string{},
			CorrectAnswer: 0,
			Explanation:   "Missing options.",
		},
	}

	if err := ValidateQuestions(questions); err == nil {
		t.Fatal("expected error when one question is invalid, got nil")
	}
}

func TestValidateQuestions_EmptySlice(t *testing.T) {
	var questions []model.LLMQuestion

	if err := ValidateQuestions(questions); err == nil {
		t.Fatal("expected error for empty question slice, got nil")
	}
}
