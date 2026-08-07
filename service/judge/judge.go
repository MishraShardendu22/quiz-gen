package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/MishraShardendu22/quiz-gen/util"
)

// sends existing questions and newly generated questions to LLM as Judge
// returns the 0-based indices of newly generated questions that are semantically equivalent to any existing question
func JudgeDuplicates(ctx context.Context, client *openrouter.Client, existing []model.Question, newQuestions []model.LLMQuestion) ([]int, *openrouter.Usage, error) {
	if len(newQuestions) == 0 {
		return []int{}, nil, nil
	}

	if len(existing) == 0 {
		return []int{}, nil, nil
	}

	var existingBuilder strings.Builder
	for i, q := range existing {
		existingBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, q.Question))
	}

	var newBuilder strings.Builder
	for i, q := range newQuestions {
		newBuilder.WriteString(fmt.Sprintf("[%d] %s\n", i, q.Question))
	}

	prompt := fmt.Sprintf(JudgePromptTemplate, existingBuilder.String(), newBuilder.String())

	respStr, usage, err := client.GenerateQuestions(ctx, prompt)
	if err != nil {
		return nil, usage, fmt.Errorf("judge LLM call failed: %w", err)
	}

	cleaned := openrouter.CleanJSON(respStr)

	var res model.JudgeResult
	if err := json.Unmarshal([]byte(cleaned), &res); err != nil {
		util.Error("failed to parse judge response", "error", err.Error())
		return []int{}, usage, nil
	}

	// Filter valid indices
	validDuplicates := make([]int, 0)
	for _, idx := range res.Duplicates {
		if idx >= 0 && idx < len(newQuestions) {
			validDuplicates = append(validDuplicates, idx)
		}
	}

	return validDuplicates, usage, nil
}
