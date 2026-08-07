package openrouter

import "fmt"

// ModelPricing contains pricing information for a model
type ModelPricing struct {
	PromptTokenPrice     float64 // Price per 1K prompt tokens
	CompletionTokenPrice float64 // Price per 1K completion tokens
}

// GetModelPricing returns pricing for a given model
// Returns (promptPrice, completionPrice) per 1K tokens
func GetModelPricing(model string) ModelPricing {
	// Add more models as they are added to the system
	switch model {
	case "inclusionai/ling-3.0-flash:free":
		return ModelPricing{
			PromptTokenPrice:     0.0,
			CompletionTokenPrice: 0.0,
		}
	default:
		// Default to free if model not found
		return ModelPricing{
			PromptTokenPrice:     0.0,
			CompletionTokenPrice: 0.0,
		}
	}
}

// calculates cost from a Usage struct
func EstimateCostFromUsage(model string, usage *Usage) float64 {
	if usage == nil {
		return 0.0
	}
	return EstimateCost(model, usage.PromptTokens, usage.CompletionTokens)
}

// calculates the estimated cost for a generation
// takes model name, prompt tokens, and completion tokens, eturns cost in USD as a float64
func EstimateCost(model string, promptTokens, completionTokens int) float64 {
	pricing := GetModelPricing(model)
	promptCost := float64(promptTokens) * pricing.PromptTokenPrice / 1000.0
	completionCost := float64(completionTokens) * pricing.CompletionTokenPrice / 1000.0
	return promptCost + completionCost
}

// returns a formatted cost string
func FormatCost(cost float64) string {
	if cost == 0.0 {
		return "$0.00"
	}
	return fmt.Sprintf("$%.6f", cost)
}
