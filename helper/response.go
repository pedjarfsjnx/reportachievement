package helper

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// 200 OK / 201 Created
func Success(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// 400 Bad Request / 401 Unauthorized / 500 Internal Server Error
func Error(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(APIResponse{
		Status:  "error",
		Message: message,
		Error:   message,
	})
}
