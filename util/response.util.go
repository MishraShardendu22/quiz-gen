package util

import (
	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/gofiber/fiber/v2"
)

/*
	Utils for sending JSON responses and error responses in a consistent format across the application.
	JSONResponse sends a successful JSON response with the provided status, message, and data.
	ErrorResponse sends an error JSON response with the provided status, message, and optional error details.

	Data - can be of any type, and the response will be wrapped in a generic Response struct.
*/

func JSONResponse[T any](c *fiber.Ctx, status int, message string, data T) error {
	return c.Status(status).JSON(model.Response[T]{
		Code:    status,
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, status int, message string, err error) error {
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
