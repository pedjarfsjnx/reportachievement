package postgre

import (
	// --- FIX IMPORT DI SINI ---
	// Mengarah ke folder root "middleware", BUKAN "app/middleware"
	"reportachievement/middleware"

	"reportachievement/app/repository/postgre" // Import untuk Filter Struct (Modul 8)
	"reportachievement/app/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterAchievementRoutes(app *fiber.App, achievementService *service.AchievementService) {
	api := app.Group("/api/v1/achievements")

	// 1. Endpoint CREATE (Modul 7)
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

	// 2. Endpoint GET ALL / LIST (Modul 8)
	api.Get("/", middleware.Protected(), func(c *fiber.Ctx) error {
		// Ambil Role dari Token (untuk logic authorization nanti)
		// role := c.Locals("role").(string)

		// Parse Query Params untuk Pagination & Filter
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "10"))
		status := c.Query("status") // e.g. ?status=draft

		// Siapkan filter
		filter := postgre.AchievementFilter{
			Page:   page,
			Limit:  limit,
			Status: status,
			// Jika nanti ingin filter berdasarkan user login (untuk mahasiswa),
			// kita bisa tambahkan StudentID di sini.
		}

		// Panggil Service GetAll
		data, total, err := achievementService.GetAll(c.Context(), filter)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Return JSON Response
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
}
