package loader

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/google/uuid"
)

// LoadContent discovers topics and documents from filesystem with content hashes
// Returns topics grouped with their documents, expressing ownership naturally
func LoadContent(contentPackDir string) ([]model.LoadedTopic, error) {
	util.Info("starting content discovery", "directory", contentPackDir)

	var loadedTopics []model.LoadedTopic
	topicsMap := make(map[string]*model.LoadedTopic)

	entries, err := os.ReadDir(contentPackDir)
	if err != nil {
		return nil, fmt.Errorf("read content-pack: %w", err)
	}

	// Discover topics (first-level directories) and their documents
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		topicPath := filepath.Join(contentPackDir, entry.Name())
		loadedTopic := &model.LoadedTopic{
			Name:      entry.Name(),
			Documents: []model.Document{},
		}

		util.Info("discovering topic", "name", entry.Name())

		if err := walkMarkdownFiles(topicPath, topicPath, &loadedTopic.Documents); err != nil {
			return nil, err
		}

		util.Info("topic discovery complete", "name", entry.Name(), "documents", len(loadedTopic.Documents))

		topicsMap[entry.Name()] = loadedTopic
		loadedTopics = append(loadedTopics, *loadedTopic)
	}

	util.Info("content discovery complete", "topics", len(loadedTopics), "total_documents", countTotalDocuments(loadedTopics))

	return loadedTopics, nil
}

// countTotalDocuments returns the total number of documents across all topics
func countTotalDocuments(topics []model.LoadedTopic) int {
	count := 0
	for _, t := range topics {
		count += len(t.Documents)
	}
	return count
}

// walkMarkdownFiles recursively finds markdown files and computes their hashes
func walkMarkdownFiles(topicRoot, dir string, documents *[]model.Document) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if err := walkMarkdownFiles(topicRoot, fullPath, documents); err != nil {
				return err
			}
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", fullPath, err)
		}

		relPath, err := filepath.Rel(topicRoot, fullPath)
		if err != nil {
			return fmt.Errorf("compute relative path %s: %w", fullPath, err)
		}

		// Compute SHA256 hash of content
		hash := sha256.Sum256(content)
		hashHex := fmt.Sprintf("%x", hash)

		*documents = append(*documents, model.Document{
			ID:      uuid.Must(uuid.NewV7()),
			Name:    entry.Name(),
			Path:    relPath,
			Content: string(content),
			Hash:    hashHex,
		})
	}

	return nil
}