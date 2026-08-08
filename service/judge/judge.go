package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/MishraShardendu22/quiz-gen/util"
)

// sends existing questions and newly generated questions to LLM as Judge
// returns the 0-based indices of newly generated questions that are semantically equivalent to any existing question
const MaxJudgeRetries = 2

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

	var totalUsage *openrouter.Usage
	var lastErr error

	for attempt := 1; attempt <= MaxJudgeRetries; attempt++ {
		if attempt > 1 {
			waitTime := time.Duration(1<<(attempt-1)) * time.Second
			util.Info("retrying judge call after backoff", "attempt", attempt, "wait_time", waitTime.String())
			select {
			case <-ctx.Done():
				return nil, totalUsage, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		respStr, usage, err := client.GenerateQuestions(ctx, prompt)
		if usage != nil {
			if totalUsage == nil {
				totalUsage = &openrouter.Usage{}
			}
			totalUsage.PromptTokens += usage.PromptTokens
			totalUsage.CompletionTokens += usage.CompletionTokens
			totalUsage.TotalTokens += usage.TotalTokens
		}

		if err != nil {
			util.Error("judge LLM call failed", "attempt", attempt, "error", err.Error())
			lastErr = fmt.Errorf("judge LLM call failed: %w", err)
			continue
		}

		cleaned := openrouter.CleanJSON(respStr)

		var res model.JudgeResult
		if err := json.Unmarshal([]byte(cleaned), &res); err != nil {
			util.Error("failed to parse judge response", "attempt", attempt, "error", err.Error())
			lastErr = fmt.Errorf("failed to parse judge response: %w", err)
			continue
		}

		// Filter valid indices
		validDuplicates := make([]int, 0)
		for _, idx := range res.Duplicates {
			if idx >= 0 && idx < len(newQuestions) {
				validDuplicates = append(validDuplicates, idx)
			}
		}

		return validDuplicates, totalUsage, nil
	}

	return nil, totalUsage, fmt.Errorf("duplicate judge failed after %d attempts: %w", MaxJudgeRetries, lastErr)
}
