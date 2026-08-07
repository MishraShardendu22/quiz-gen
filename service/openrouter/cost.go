package openrouter

import (
	"fmt"

	"github.com/MishraShardendu22/quiz-gen/util"
)

// ModelPricing contains pricing information for a model
type ModelPricing struct {
	PromptTokenPrice     float64 // Price per 1K prompt tokens
	CompletionTokenPrice float64 // Price per 1K completion tokens
}

// GetModelPricing returns pricing for a given model
// Returns (promptPrice, completionPrice) per 1K tokens
func GetModelPricing(model string) ModelPricing {
	switch model {
	case util.Config.ModelName:
		return ModelPricing{
			PromptTokenPrice:     0.0,
			CompletionTokenPrice: 0.0,
		}
	default:
		return ModelPricing{
			PromptTokenPrice:     0.0,
			CompletionTokenPrice: 0.0,
		}
	}
}

// CalculateCostFromUsage calculates actual cost from a Usage struct
func CalculateCostFromUsage(model string, usage *Usage) float64 {
	if usage == nil {
		return 0.0
	}
	return CalculateCost(model, usage.PromptTokens, usage.CompletionTokens)
}

// CalculateCost calculates the cost for actual reported tokens returned by OpenRouter
func CalculateCost(model string, promptTokens, completionTokens int) float64 {
	pricing := GetModelPricing(model)
	promptCost := float64(promptTokens) * pricing.PromptTokenPrice / 1000.0
	completionCost := float64(completionTokens) * pricing.CompletionTokenPrice / 1000.0
	return promptCost + completionCost
}

// FormatCost returns a formatted cost string
func FormatCost(cost float64) string {
	if cost == 0.0 {
		return "$0.00"
	}
	return fmt.Sprintf("$%.6f", cost)
}
