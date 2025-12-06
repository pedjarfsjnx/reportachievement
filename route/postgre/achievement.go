package postgre

import (
	"reportachievement/app/repository/postgre"
	"reportachievement/app/service"
	"reportachievement/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterAchievementRoutes(app *fiber.App, achievementService *service.AchievementService) {
	api := app.Group("/api/v1/achievements")

	// 1. CREATE
	api.Post("/", middleware.Protected(), func(c *fiber.Ctx) error {
		userIDStr := c.Locals("user_id").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
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

	// 2. GET LIST
	api.Get("/", middleware.Protected(), func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		status := c.Query("status")

		filter := postgre.AchievementFilter{
			Page:   page,
			Limit:  limit,
			Status: status,
		}

		data, total, err := achievementService.GetAll(c.Context(), filter)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   data,
			"meta": fiber.Map{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
	})

	// 3. DELETE (BARU - MODUL 9)
	api.Delete("/:id", middleware.Protected(), func(c *fiber.Ctx) error {
		userIDStr := c.Locals("user_id").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
		}

		// Ambil ID Prestasi dari URL Parameter
		achIDStr := c.Params("id")
		achID, err := uuid.Parse(achIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Achievement ID UUID"})
		}

		// Panggil Service Delete
		if err := achievementService.Delete(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Achievement soft deleted successfully",
		})
	})
}
