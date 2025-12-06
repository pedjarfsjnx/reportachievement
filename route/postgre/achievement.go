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

	// 3. DELETE (Soft Delete)
	api.Delete("/:id", middleware.Protected(), func(c *fiber.Ctx) error {
		userIDStr := c.Locals("user_id").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
		}

		achIDStr := c.Params("id")
		achID, err := uuid.Parse(achIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Achievement ID UUID"})
		}

		if err := achievementService.Delete(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Achievement soft deleted successfully",
		})
	})

	// --- BARU (MODUL 10): WORKFLOW ROUTES ---

	// 4. SUBMIT (Mahasiswa)
	api.Post("/:id/submit", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, _ := uuid.Parse(c.Locals("user_id").(string))
		achID, _ := uuid.Parse(c.Params("id"))

		if err := achievementService.Submit(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success", "message": "Achievement submitted"})
	})

	// 5. VERIFY (Dosen)
	api.Post("/:id/verify", middleware.Protected(), func(c *fiber.Ctx) error {
		// Cek Role (Simple Check)
		role := c.Locals("role").(string)
		if role != "Dosen Wali" && role != "Admin" { // Admin boleh bantu verify utk testing
			return c.Status(403).JSON(fiber.Map{"error": "Only Lecturer can verify"})
		}

		userID, _ := uuid.Parse(c.Locals("user_id").(string))
		achID, _ := uuid.Parse(c.Params("id"))

		if err := achievementService.Verify(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success", "message": "Achievement verified"})
	})

	// 6. REJECT (Dosen)
	api.Post("/:id/reject", middleware.Protected(), func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "Dosen Wali" && role != "Admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Only Lecturer can reject"})
		}

		userID, _ := uuid.Parse(c.Locals("user_id").(string))
		achID, _ := uuid.Parse(c.Params("id"))

		// Ambil Note dari Body
		var req struct {
			Note string `json:"note"`
		}
		if err := c.BodyParser(&req); err != nil || req.Note == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
		}

		if err := achievementService.Reject(c.Context(), userID, achID, req.Note); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success", "message": "Achievement rejected"})
	})
}
