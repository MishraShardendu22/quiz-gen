package llmretry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/MishraShardendu22/quiz-gen/service/validator"
	"github.com/MishraShardendu22/quiz-gen/util"
)

const MaxRetries = 2

// attempts to generate questions with LLM retries
// Flow: Generate -> Clean -> Parse -> Validate -> Success or Retry (Max 3 attempts)
func GenerateWithRetry(ctx context.Context, client *openrouter.Client, sessionID string, prompt string) ([]model.LLMQuestion, *openrouter.Usage, error) {
	var lastErr error
	var totalUsage *openrouter.Usage

	for attempt := 1; attempt <= MaxRetries; attempt++ {
		util.Info("LLM generation attempt", "session_id", sessionID, "attempt", attempt, "max_retries", MaxRetries)

		// 1. Generate
		response, usage, err := client.GenerateQuestions(ctx, prompt)
		if err != nil {
			util.Error("generation failed", "session_id", sessionID, "attempt", attempt, "retry_reason", err.Error())
			lastErr = fmt.Errorf("generation failed: %w", err)
			continue
		}

		// earlier we were not tracking total usage across retries, now we are tracking it
		if usage != nil {
			if totalUsage == nil {
				totalUsage = &openrouter.Usage{}
			}
			totalUsage.PromptTokens += usage.PromptTokens
			totalUsage.CompletionTokens += usage.CompletionTokens
			totalUsage.TotalTokens += usage.TotalTokens
		}

		// 2. Clean JSON
		cleaned := openrouter.CleanJSON(response)

		// 3. Parse JSON - if parsing fails, we retry
		var llmResp model.LLMResponse
		if err := json.Unmarshal([]byte(cleaned), &llmResp); err != nil {
			util.Error("JSON parsing failed", "session_id", sessionID, "attempt", attempt, "retry_reason", err.Error())
			lastErr = fmt.Errorf("JSON parsing failed: %w", err)
			continue
		}

		// 4. Validate questions
		if err := validator.ValidateQuestions(llmResp.Questions); err != nil {
			util.Error("validation failed", "session_id", sessionID, "attempt", attempt, "validation_failures", err.Error(), "retry_reason", err.Error())
			lastErr = fmt.Errorf("validation failed: %w", err)
			continue
		}

		// 5. Success
		util.Info("generation success", "session_id", sessionID, "attempt", attempt, "question_count", len(llmResp.Questions))
		return llmResp.Questions, totalUsage, nil
	}

	return nil, totalUsage, fmt.Errorf("failed to generate valid questions after %d attempts: %v", MaxRetries, lastErr)
}
