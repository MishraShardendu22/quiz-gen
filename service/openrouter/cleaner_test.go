package openrouter

import (
	"testing"
)

func TestCleanJSON_WhitespaceNormalization(t *testing.T) {
	input := "   {\"name\": \"quiz\"}   \n\n"
	expected := "{\"name\": \"quiz\"}"

	result := CleanJSON(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCleanJSON_MultipleBlankLines(t *testing.T) {
	input := "```json\n{\n\n  \"question\": \"What is Go?\"\n\n}\n```"
	expected := "{\n\n  \"question\": \"What is Go?\"\n\n}"

	result := CleanJSON(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCleanJSON_Tabs(t *testing.T) {
	input := "```json\n\t{\n\t\t\"key\":\t\"value\"\n\t}\n```"
	expected := "{\n\t\t\"key\":\t\"value\"\n\t}"

	result := CleanJSON(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCleanJSON_EmptyString(t *testing.T) {
	input := "   "
	expected := ""

	result := CleanJSON(input)
	if result != expected {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestCleanJSON_UnicodeHandling(t *testing.T) {
	input := "```json\n{\"text\": \"Hello 🚀 世界 (Fire Safety)\"}\n```"
	expected := "{\"text\": \"Hello 🚀 世界 (Fire Safety)\"}"

	result := CleanJSON(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
