package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadContent_Success(t *testing.T) {
	root := t.TempDir()

	topicDir := filepath.Join(root, "safety")
	err := os.MkdirAll(topicDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	content := "# Fire Safety\n\nFire extinguishers should be inspected monthly."
	filePath := filepath.Join(topicDir, "fire.md")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	topics, err := LoadContent(root)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	topic := topics[0]
	if topic.Name != "safety" {
		t.Errorf("expected topic name 'safety', got %q", topic.Name)
	}

	if len(topic.Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(topic.Documents))
	}

	doc := topic.Documents[0]
	if doc.Name != "fire.md" {
		t.Errorf("expected document name 'fire.md', got %q", doc.Name)
	}
	if doc.Path != "fire.md" {
		t.Errorf("expected path 'fire.md', got %q", doc.Path)
	}
	if doc.Content != content {
		t.Errorf("expected content %q, got %q", content, doc.Content)
	}
	if doc.Hash == "" {
		t.Error("expected non-empty document hash")
	}
}

func TestLoadContent_IgnoresNonMarkdownAndHiddenFiles(t *testing.T) {
	root := t.TempDir()

	topicDir := filepath.Join(root, "go")
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		t.Fatal(err)
	}

	hiddenDir := filepath.Join(root, ".git")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Root"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(topicDir, "intro.md"), []byte("# Intro"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(topicDir, "notes.txt"), []byte("some notes"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(topicDir, ".draft.md"), []byte("# Draft"), 0644); err != nil {
		t.Fatal(err)
	}

	topics, err := LoadContent(root)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	if len(topics[0].Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(topics[0].Documents))
	}

	if topics[0].Documents[0].Name != "intro.md" {
		t.Errorf("expected document 'intro.md', got %q", topics[0].Documents[0].Name)
	}
}

func TestLoadContent_NestedMarkdownFiles(t *testing.T) {
	root := t.TempDir()

	nestedDir := filepath.Join(root, "cpp", "stl")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(nestedDir, "vector.md"), []byte("# Vector"), 0644); err != nil {
		t.Fatal(err)
	}

	topics, err := LoadContent(root)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	if len(topics[0].Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(topics[0].Documents))
	}

	doc := topics[0].Documents[0]
	if doc.Name != "vector.md" {
		t.Errorf("expected doc name 'vector.md', got %q", doc.Name)
	}
	expectedPath := filepath.Join("stl", "vector.md")
	if doc.Path != expectedPath {
		t.Errorf("expected relative path %q, got %q", expectedPath, doc.Path)
	}
}

func TestLoadContent_NonExistentDirectory(t *testing.T) {
	_, err := LoadContent("/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}
