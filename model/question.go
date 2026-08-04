package model

import "github.com/google/uuid"

// Question represents a generated MCQ
type Question struct {
	ID            uuid.UUID `json:"id"`
	SessionID     uuid.UUID `json:"session_id"`
	Question      string    `json:"question"`
	Option1       string    `json:"option_1"`
	Option2       string    `json:"option_2"`
	Option3       string    `json:"option_3"`
	Option4       string    `json:"option_4"`
	CorrectAnswer int       `json:"correct_answer"` // 0-3
	Explanation   string    `json:"explanation"`
	CreatedAt     int64     `json:"created_at"`
}

// LLMQuestion represents a single question from the LLM response
type LLMQuestion struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
}

// LLMResponse represents the structure of the LLM response
type LLMResponse struct {
	Questions []LLMQuestion `json:"questions"`
}
