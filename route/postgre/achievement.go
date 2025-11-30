package postgre

import (
	// FIX IMPORT PATH: Dari "app/middleware" ke "middleware" (root)
	"reportachievement/app/service"
	"reportachievement/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterAchievementRoutes(app *fiber.App, achievementService *service.AchievementService) {
	api := app.Group("/api/v1/achievements")

	// Panggil middleware.Protected() dari package yang baru
	api.Post("/", middleware.Protected(), func(c *fiber.Ctx) error {
		userIDStr := c.Locals("user_id").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID in token"})
		}

		var req service.CreateAchievementRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
		}

		result, err := achievementService.Create(c.Context(), userID, req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(201).JSON(fiber.Map{
			"status":  "success",
			"message": "Achievement draft created",
			"data":    result,
		})
	})
}
