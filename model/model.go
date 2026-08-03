package model

import (
	"github.com/google/uuid"
)

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
