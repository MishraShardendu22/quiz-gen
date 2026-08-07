package loader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitByHeading_H1AndH2Headings(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "safety", "fire.md")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatal(err)
	}

	content := `# Fire Safety

Fire extinguishers should be inspected monthly.

## Types

Water extinguishers

CO₂ extinguishers

## Inspection

Inspect pressure gauge monthly.`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitByHeading(string(data))
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if !strings.HasPrefix(chunks[0], "# Fire Safety") {
		t.Errorf("chunk 0 should start with H1 heading, got %q", chunks[0])
	}
	if !strings.HasPrefix(chunks[1], "## Types") {
		t.Errorf("chunk 1 should start with '## Types', got %q", chunks[1])
	}
	if !strings.HasPrefix(chunks[2], "## Inspection") {
		t.Errorf("chunk 2 should start with '## Inspection', got %q", chunks[2])
	}
}

func TestSplitByHeading_PreservesH3WithinH2Chunk(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "doc.md")

	content := `## Types

### Portable Extinguishers

Water extinguishers

### Fixed Systems

Sprinklers`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitByHeading(string(data))
	if len(chunks) != 1 {
		t.Fatalf("expected H3 headings to remain in single H2 chunk, got %d chunks", len(chunks))
	}

	if !strings.Contains(chunks[0], "### Portable Extinguishers") || !strings.Contains(chunks[0], "### Fixed Systems") {
		t.Errorf("chunk should preserve H3 sub-headings, got %q", chunks[0])
	}
}

func TestSplitByHeading_NoHeadings(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "plain.md")

	content := "This is a plain document without any markdown headings."
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitByHeading(string(data))
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != content {
		t.Errorf("expected content %q, got %q", content, chunks[0])
	}
}

func TestSplitByHeading_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.md")

	if err := os.WriteFile(filePath, []byte("   \n\n   "), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	chunks := SplitByHeading(string(data))
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty content, got %d", len(chunks))
	}
}
