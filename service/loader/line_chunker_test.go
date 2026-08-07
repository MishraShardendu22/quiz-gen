package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitClosestTo500Lines_Under500Lines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "sample.txt")

	lines := make([]string, 10)
	for i := 0; i < 10; i++ {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	content := strings.Join(lines, "\n")

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitClosestTo500Lines(string(data))
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != content {
		t.Errorf("expected content %q, got %q", content, chunks[0])
	}
}

func TestSplitClosestTo500Lines_Exceeds500Lines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "large.txt")

	lines := make([]string, 550)
	for i := 0; i < 550; i++ {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	content := strings.Join(lines, "\n")

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitClosestTo500Lines(string(data))
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	firstChunkLines := strings.Split(chunks[0], "\n")
	if len(firstChunkLines) != 500 {
		t.Errorf("expected 500 lines in first chunk, got %d", len(firstChunkLines))
	}

	secondChunkLines := strings.Split(chunks[1], "\n")
	if len(secondChunkLines) != 50 {
		t.Errorf("expected 50 lines in second chunk, got %d", len(secondChunkLines))
	}
}

func TestSplitClosestTo500Lines_Exact500Lines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "exact.txt")

	lines := make([]string, 500)
	for i := 0; i < 500; i++ {
		lines[i] = fmt.Sprintf("Line %d", i+1)
	}
	content := strings.Join(lines, "\n")

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitClosestTo500Lines(string(data))
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	resultLines := strings.Split(chunks[0], "\n")
	if len(resultLines) != 500 {
		t.Errorf("expected 500 lines, got %d", len(resultLines))
	}
}

func TestSplitClosestTo500Lines_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.txt")

	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitClosestTo500Lines(string(data))
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty content, got %d", len(chunks))
	}
}
