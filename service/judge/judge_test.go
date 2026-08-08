package judge

import (
	"context"
	"testing"

	"github.com/MishraShardendu22/quiz-gen/model"
)

func TestJudgeDuplicates_EmptyQuestions(t *testing.T) {
	ctx := context.Background()

	dup, usage, err := JudgeDuplicates(ctx, nil, nil, nil)

	// no error should be there for empty questions
	if err != nil {
		t.Fatalf("expected no error for empty questions, got %v", err)
	}

	// length of duplicates should be 0 for empty questions
	if len(dup) != 0 {
		t.Fatalf("expected 0 duplicates for empty questions, got %d", len(dup))
	}

	// usage should be nil for empty questions
	if usage != nil {
		t.Fatalf("expected nil usage for empty questions, got %v", usage)
	}

	dup, usage, err = JudgeDuplicates(ctx, nil, []model.Question{{Question: "Existing?"}}, nil)

	// no error should be there for empty new questions
	if err != nil {
		t.Fatalf("expected no error for empty new questions, got %v", err)
	}

	// length of duplicates should be 0 for empty new questions
	if len(dup) != 0 {
		t.Fatalf("expected 0 duplicates, got %d", len(dup))
	}
}
