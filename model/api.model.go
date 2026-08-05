package model

type Response[T any] struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type GenerateRequest struct {
	TopicID        string `json:"topic_id"`
	RequestedCount int    `json:"requested_count"`
	TokenBudget    int    `json:"token_budget"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type GenerateResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

type RetrySessionRequest struct {
	TokenBudget int `json:"token_budget"`
}
