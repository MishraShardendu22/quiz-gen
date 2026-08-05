package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/loader"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// SyncTopicsDocumentsChunks synchronizes filesystem content with database incrementally
// Iterates over loaded topics, resolves/creates each topic's database UUID,
// then processes its documents with that UUID
func SyncTopicsDocumentsChunks(db *sql.DB, loadedTopics []model.LoadedTopic) error {
	start := time.Now()
	util.Info("starting content synchronization", "topics", len(loadedTopics))

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	filesystemDocs := make(map[string]bool)
	stats := &syncStats{}

	// Process each topic and its documents
	for _, loadedTopic := range loadedTopics {
		util.Info("syncing topic", "name", loadedTopic.Name, "documents", len(loadedTopic.Documents))

		// Look up or create topic
		var topicID string
		err := tx.QueryRow("SELECT id FROM topics WHERE name = ?", loadedTopic.Name).Scan(&topicID)

		if err == sql.ErrNoRows {
			topicID = uuid.Must(uuid.NewV7()).String()
			if _, err := tx.Exec(
				"INSERT INTO topics (id, name, status) VALUES (?, ?, ?)",
				topicID, loadedTopic.Name, "pending",
			); err != nil {
				return fmt.Errorf("insert topic %s: %w", loadedTopic.Name, err)
			}
			util.Info("created new topic", "name", loadedTopic.Name, "id", topicID)
			stats.topicsCreated++

		} else if err != nil {
			return fmt.Errorf("query topic %s: %w", loadedTopic.Name, err)
		} else {
			util.Info("using existing topic", "name", loadedTopic.Name, "id", topicID)
			stats.topicsExisting++
		}

		// Process this topic's documents with its resolved UUID
		for _, doc := range loadedTopic.Documents {
			key := topicID + "|" + doc.Path
			filesystemDocs[key] = true

			// Check if document already exists
			var existingID, existingHash string
			err := tx.QueryRow(
				"SELECT id, content_hash FROM documents WHERE topic_id = ? AND path = ?",
				topicID, doc.Path,
			).Scan(&existingID, &existingHash)

			if err == sql.ErrNoRows {
				// New document: insert it
				newID := uuid.Must(uuid.NewV7()).String()
				if _, err := tx.Exec(
					"INSERT INTO documents (id, topic_id, name, path, content_hash) VALUES (?, ?, ?, ?, ?)",
					newID, topicID, doc.Name, doc.Path, doc.Hash,
				); err != nil {
					return fmt.Errorf("insert document %s: %w", doc.Path, err)
				}

				// Chunk and store
				chunkCount, err := insertChunks(tx, topicID, newID, doc.Content)
				if err != nil {
					return err
				}

				util.Info("created new document", "topic", loadedTopic.Name, "path", doc.Path, "chunks", chunkCount)
				stats.documentsCreated++
				stats.chunksCreated += chunkCount

			} else if err != nil {
				return fmt.Errorf("query document %s: %w", doc.Path, err)
			} else {
				// Document exists: check hash
				if existingHash == doc.Hash {
					// Hash unchanged: skip completely
					util.Info("skipping unchanged document", "topic", loadedTopic.Name, "path", doc.Path)
					stats.documentsSkipped++
					continue
				}

				// Hash changed: delete old chunks, update document, re-chunk
				if _, err := tx.Exec("DELETE FROM chunks WHERE document_id = ?", existingID); err != nil {
					return fmt.Errorf("delete old chunks for %s: %w", doc.Path, err)
				}

				if _, err := tx.Exec(
					"UPDATE documents SET name = ?, content_hash = ? WHERE id = ?",
					doc.Name, doc.Hash, existingID,
				); err != nil {
					return fmt.Errorf("update document %s: %w", doc.Path, err)
				}

				chunkCount, err := insertChunks(tx, topicID, existingID, doc.Content)
				if err != nil {
					return err
				}

				util.Info("updated document", "topic", loadedTopic.Name, "path", doc.Path, "chunks", chunkCount)
				stats.documentsUpdated++
				stats.chunksCreated += chunkCount
			}
		}
	}

	// Delete documents that exist in DB but not on filesystem
	allDBDocs, err := tx.Query("SELECT id, topic_id, path FROM documents")
	if err != nil {
		return fmt.Errorf("query all documents: %w", err)
	}

	type orphanDoc struct {
		id      string
		topicID string
		path    string
	}
	var orphans []orphanDoc

	for allDBDocs.Next() {
		var id, topicID, path string
		if err := allDBDocs.Scan(&id, &topicID, &path); err != nil {
			allDBDocs.Close()
			return err
		}

		key := topicID + "|" + path
		if !filesystemDocs[key] {
			orphans = append(orphans, orphanDoc{id: id, topicID: topicID, path: path})
		}
	}
	if err := allDBDocs.Err(); err != nil {
		allDBDocs.Close()
		return err
	}
	allDBDocs.Close()

	for _, doc := range orphans {
		// Document not on filesystem: delete its chunks and the document
		if _, err := tx.Exec("DELETE FROM chunks WHERE document_id = ?", doc.id); err != nil {
			return fmt.Errorf("delete chunks for orphaned doc: %w", err)
		}
		if _, err := tx.Exec("DELETE FROM documents WHERE id = ?", doc.id); err != nil {
			return fmt.Errorf("delete orphaned document: %w", err)
		}
		util.Info("deleted orphaned document", "path", doc.path, "id", doc.id)
		stats.documentsDeleted++
	}

	// Mark all topics as completed
	if _, err := tx.Exec("UPDATE topics SET status = ? WHERE status != ?", "completed", "completed"); err != nil {
		return fmt.Errorf("update topic status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	elapsed := time.Since(start)
	util.Info("synchronization complete",
		"duration_ms", elapsed.Milliseconds(),
		"topics_created", stats.topicsCreated,
		"topics_existing", stats.topicsExisting,
		"documents_created", stats.documentsCreated,
		"documents_updated", stats.documentsUpdated,
		"documents_unchanged", stats.documentsSkipped,
		"documents_deleted", stats.documentsDeleted,
		"chunks_created", stats.chunksCreated,
	)

	return nil
}

// syncStats tracks synchronization metrics
type syncStats struct {
	topicsCreated    int
	topicsExisting   int
	documentsCreated int
	documentsUpdated int
	documentsSkipped int
	documentsDeleted int
	chunksCreated    int
}

// insertChunks chunks a document and inserts chunks into database
// Returns the number of chunks created
func insertChunks(tx *sql.Tx, topicID, documentID, content string) (int, error) {
	chunks := loader.SplitByHeading(content)
	for i, chunk := range chunks {
		if _, err := tx.Exec(
			"INSERT INTO chunks (id, topic_id, document_id, chunk_index, content) VALUES (?, ?, ?, ?, ?)",
			uuid.Must(uuid.NewV7()).String(),
			topicID,
			documentID,
			i,
			chunk,
		); err != nil {
			return 0, fmt.Errorf("insert chunk: %w", err)
		}
	}
	return len(chunks), nil
}
