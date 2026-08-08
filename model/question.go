package model

import "github.com/google/uuid"

/*
	Question represents a generated MCQ stored in database
	LLMQuestion represents a single question from the LLM response
	LLMResponse represents the structure of the LLM response
	JudgeResult represents the response from the LLM Judge
*/

type Question struct {
	ID            uuid.UUID `json:"id"`
	SessionID     uuid.UUID `json:"session_id"`
	Question      string    `json:"question"`
	Option1       string    `json:"option_1"`
	Option2       string    `json:"option_2"`
	Option3       string    `json:"option_3"`
	Option4       string    `json:"option_4"`
	CorrectAnswer int       `json:"correct_answer"`
	Explanation   string    `json:"explanation"`
	CreatedAt     int64     `json:"created_at"`
}

type LLMQuestion struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
}

type LLMResponse struct {
	Questions []LLMQuestion `json:"questions"`
}

type JudgeResult struct {
	Duplicates []int `json:"duplicates"`
}
