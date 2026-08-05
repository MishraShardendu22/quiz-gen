package model

type SessionUsage struct {
	SessionID        string  `json:"session_id"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

type UsageBreakdown struct {
	SessionID        string  `json:"session_id"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

type UsageResponse struct {
	TotalPromptTokens     int              `json:"total_prompt_tokens"`
	TotalCompletionTokens int              `json:"total_completion_tokens"`
	TotalTokens           int              `json:"total_tokens"`
	EstimatedCost         float64          `json:"estimated_cost"`
	Sessions              []UsageBreakdown `json:"sessions"`
}
