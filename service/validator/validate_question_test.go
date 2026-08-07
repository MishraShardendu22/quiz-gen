package validator

import (
	"testing"

	"github.com/MishraShardendu22/quiz-gen/model"
)

func TestValidateQuestion_Valid(t *testing.T) {
	q := model.LLMQuestion{
		Question:      "How often should fire extinguishers be inspected?",
		Options:       []string{"Daily", "Weekly", "Monthly", "Yearly"},
		CorrectAnswer: 2,
		Explanation:   "Inspect pressure gauge monthly per safety guidelines.",
	}

	if err := ValidateQuestion(q); err != nil {
		t.Fatalf("expected valid question to pass validation, got: %v", err)
	}
}

func TestValidateQuestion_EmptyQuestion(t *testing.T) {
	q := model.LLMQuestion{
		Question:      "   ",
		Options:       []string{"Option A", "Option B", "Option C", "Option D"},
		CorrectAnswer: 0,
		Explanation:   "Valid explanation.",
	}

	if err := ValidateQuestion(q); err == nil {
		t.Fatal("expected error for empty question text, got nil")
	}
}

func TestValidateQuestion_FewerThanFourOptions(t *testing.T) {
	q := model.LLMQuestion{
		Question:      "What is the chemical symbol for water?",
		Options:       []string{"H2O", "CO2", "O2"},
		CorrectAnswer: 0,
		Explanation:   "Water is composed of hydrogen and oxygen.",
	}

	if err := ValidateQuestion(q); err == nil {
		t.Fatal("expected error for fewer than 4 options, got nil")
	}
}

func TestValidateQuestion_InvalidCorrectAnswer(t *testing.T) {
	q := model.LLMQuestion{
		Question:      "What is the main component of air?",
		Options:       []string{"Oxygen", "Nitrogen", "Argon", "Carbon Dioxide"},
		CorrectAnswer: 5,
		Explanation:   "Nitrogen makes up about 78% of Earth's atmosphere.",
	}

	if err := ValidateQuestion(q); err == nil {
		t.Fatal("expected error for out-of-range correct answer index, got nil")
	}
}

func TestValidateQuestion_EmptyExplanation(t *testing.T) {
	q := model.LLMQuestion{
		Question:      "Which planet is known as the Red Planet?",
		Options:       []string{"Venus", "Mars", "Jupiter", "Saturn"},
		CorrectAnswer: 1,
		Explanation:   "   ",
	}

	if err := ValidateQuestion(q); err == nil {
		t.Fatal("expected error for empty explanation, got nil")
	}
}
