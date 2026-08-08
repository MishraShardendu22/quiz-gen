package model

/*
	Metrics / Usage models for the application.
	SessionUsage represents the usage metrics for a single quiz generation session.
	UsageBreakdown represents the usage metrics for a single session, including prompt and completion tokens, total tokens, and estimated cost.
	UsageResponse represents the aggregated usage metrics across all sessions, including total prompt tokens, total completion tokens, total tokens, estimated cost, and a breakdown of usage per session.
*/

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
