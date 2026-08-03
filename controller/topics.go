package controller

import (
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
)

// GetTopics retrieves all topics with document counts
func GetTopics(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(`
			SELECT t.id, t.name, t.status, COUNT(d.id) as document_count
			FROM topics t
			LEFT JOIN documents d ON t.id = d.topic_id
			GROUP BY t.id, t.name, t.status
			ORDER BY t.name
		`)
		if err != nil {
			return util.Error(c, 500, "Failed to fetch topics", err)
		}
		defer rows.Close()

		var topics []map[string]interface{}

		for rows.Next() {
			var id, name, status string
			var docCount int

			if err := rows.Scan(&id, &name, &status, &docCount); err != nil {
				return util.Error(c, 500, "Failed to fetch topics", err)
			}

			topics = append(topics, map[string]interface{}{
				"id":             id,
				"name":           name,
				"status":         status,
				"document_count": docCount,
			})
		}

		if err := rows.Err(); err != nil {
			return util.Error(c, 500, "Failed to fetch topics", err)
		}

		return util.JSON(c, 200, "Topics retrieved successfully", topics)
	}
}