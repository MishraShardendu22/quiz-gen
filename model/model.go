package model

import (
	"github.com/google/uuid"
)

/*
	Base models for the application.

	Topic represents a logical topic under which documents are grouped.
	Document represents a file belonging to a topic. The Path field stores the relative path from the topic root, preserving nested directories.

	LoadedTopic is the in-memory representation produced by the content loader.
		- It contains a topic name and every document discovered recursively under that topic.

	Example:

	content-pack/
	└── privacy/
		├── policy.md
		└── nested/
			└── access-control.md

	becomes

	LoadedTopic{
		Name: "privacy",
		Documents: []Document{
			{
				Name: "policy.md",
				Path: "policy.md",
			},
			{
				Name: "access-control.md",
				Path: "nested/access-control.md",
			},
		},
	}

	SessionStatus represents the lifecycle state of a quiz generation session
	Session represents a single quiz generation request
*/

type Topic struct {
	ID       uuid.UUID  `json:"id"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Name     string     `json:"name"`
	Status   string     `json:"status"`
}

type Document struct {
	ID      uuid.UUID `json:"id"`
	TopicID uuid.UUID `json:"topic_id"`
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Content string    `json:"content,omitempty"`
	Hash    string    `json:"hash,omitempty"`
}

type LoadedTopic struct {
	Name      string
	Documents []Document
}

type SessionStatus string

const (
	SessionPending    SessionStatus = "pending"
	SessionProcessing SessionStatus = "processing"
	SessionCompleted  SessionStatus = "completed"
	SessionFailed     SessionStatus = "failed"
)

type Session struct {
	ID             uuid.UUID     `json:"id"`
	TopicID        uuid.UUID     `json:"topic_id"`
	Status         SessionStatus `json:"status"`
	RequestedCount int           `json:"requested_count"`
	GeneratedCount int           `json:"generated_count"`
	TokenBudget    int           `json:"token_budget"`
	TokensUsed     int           `json:"tokens_used"`
	CreatedAt      int64         `json:"created_at"`
	UpdatedAt      int64         `json:"updated_at"`
	Error          *string       `json:"error,omitempty"`
	Questions      []Question    `json:"questions,omitempty"`
}
