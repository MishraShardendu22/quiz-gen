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

// load content from the content-pack directory, discovering topics and their documents
func LoadContent(contentPackDir string) ([]model.LoadedTopic, error) {
	util.Info("starting content discovery", "directory", contentPackDir)

	var loadedTopics []model.LoadedTopic
	topicsMap := make(map[string]*model.LoadedTopic)

	// find the folders in the content-pack directory
	entries, err := os.ReadDir(contentPackDir)
	if err != nil {
		return nil, fmt.Errorf("read content-pack: %w", err)
	}

	// traverse the folderes
	for _, entry := range entries {

		// if the entry is not a directory or is a hidden folder, skip it
		// .Name() returns the base name of the file or directory
		// 		//	content-pack/
		//	 ├── cpp/
		//	 │   ├── intro.md
		//	 │   └── stl/
		//	 └── python/
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// for each topic folder, discover markdown files and compute their hashes
		topicPath := filepath.Join(contentPackDir, entry.Name())

		// create a new LoadedTopic struct for the topic
		loadedTopic := &model.LoadedTopic{
			// folder name
			Name: entry.Name(),

			// initialize an empty slice of documents
			Documents: []model.Document{},
		}

		util.Info("discovering topic", "name", entry.Name())

		// walk the topic folder recursively to find markdown files and compute their hashes
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

// walkMarkdownFiles recursively finds markdown files and computes their hashes
func walkMarkdownFiles(topicRoot, dir string, documents *[]model.Document) error {
	// read the folders
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	// itreate over the entries in the directory
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// compute the full path of the entry
		fullPath := filepath.Join(dir, entry.Name())

		// if the entry is a directory, recursively walk it with the same function
		if entry.IsDir() {
			if err := walkMarkdownFiles(topicRoot, fullPath, documents); err != nil {
				return err
			}

			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// read the content of the file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", fullPath, err)
		}

		// file path relative to the topic root
		relPath, err := filepath.Rel(topicRoot, fullPath)
		if err != nil {
			return fmt.Errorf("compute relative path %s: %w", fullPath, err)
		}

		// compute SHA256 hash of content
		hash := sha256.Sum256(content)
		hashHex := fmt.Sprintf("%x", hash)

		// append the document to the documents slice
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

// countTotalDocuments returns the total number of documents across all topics (only used for logging)
func countTotalDocuments(topics []model.LoadedTopic) int {
	count := 0
	for _, t := range topics {
		count += len(t.Documents)
	}
	return count
}
