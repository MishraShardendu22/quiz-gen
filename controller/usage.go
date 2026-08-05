package controller

import (
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/service/worker"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
)

// GetUsage handles GET /usage requests
func GetUsage(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		report, err := worker.GetUsageReport(db)
		if err != nil {
			util.Error("failed to get usage report", "error", err.Error())
			return util.ErrorResponse(c, 500, "Failed to fetch usage report", err)
		}

		util.Info("usage report fetched", "total_tokens", report.TotalTokens)

		return util.JSONResponse(c, 200, "Usage report fetched successfully", report)
	}
}
