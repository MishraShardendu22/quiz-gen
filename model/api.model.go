package model

/*
	Response is a generic struct that represents a standard API response format. (non error responses)
	ErrorResponse is a struct that represents a standard API error response format.
	GenerateRequest is a struct that represents the request payload for generating a quiz.
	GenerateResponse is a struct that represents the response payload for a quiz generation request.
	RetrySessionRequest is a struct that represents the request payload for retrying a quiz generation session.
*/ 
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
