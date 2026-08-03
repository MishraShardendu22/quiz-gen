package util

import (
	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/gofiber/fiber/v2"
)

func JSON[T any](c *fiber.Ctx, status int, message string, data T) error {
	return c.Status(status).JSON(model.Response[T]{
		Code:    status,
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, message string, err error) error {
	resp := model.ErrorResponse{
		Code:    status,
		Success: false,
		Message: message,
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return c.Status(status).JSON(resp)
}
