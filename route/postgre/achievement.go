package postgre

import (
	"fmt"
	"log"
	"path/filepath"
	"reportachievement/app/repository/postgre"
	"reportachievement/app/service"
	"reportachievement/middleware"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterAchievementRoutes(app *fiber.App, achievementService *service.AchievementService) {
	api := app.Group("/api/v1/achievements")

	// Helper function: Ambil UserID dengan aman (Anti-Panic)
	getUserID := func(c *fiber.Ctx) (uuid.UUID, error) {
		claims := c.Locals("user_id")
		if claims == nil {
			return uuid.Nil, fmt.Errorf("user_id not found in token context")
		}
		// Pakai Sprintf memaksa konversi ke string apapun tipe aslinya
		return uuid.Parse(fmt.Sprintf("%v", claims))
	}

	// 1. CREATE
	api.Post("/", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: " + err.Error()})
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

	// 3. DELETE
	api.Delete("/:id", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}

		achID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Achievement ID"})
		}

		if err := achievementService.Delete(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Achievement soft deleted successfully",
		})
	})

	// 4. SUBMIT
	api.Post("/:id/submit", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		achID, _ := uuid.Parse(c.Params("id"))

		if err := achievementService.Submit(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success", "message": "Achievement submitted"})
	})

	// 5. VERIFY
	api.Post("/:id/verify", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		achID, _ := uuid.Parse(c.Params("id"))

		// Note: Idealnya cek role lagi, tapi middleware sudah handle basic auth
		if err := achievementService.Verify(c.Context(), userID, achID); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success", "message": "Achievement verified"})
	})

	// 6. REJECT
	api.Post("/:id/reject", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, err := getUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		achID, _ := uuid.Parse(c.Params("id"))

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

	// 7. UPLOAD EVIDENCE (FIXED & DEBUGGED)
	api.Post("/:id/attachments", middleware.Protected(), func(c *fiber.Ctx) error {
		log.Println("📥 [UPLOAD] Request masuk...")

		// A. Validasi User
		userID, err := getUserID(c)
		if err != nil {
			log.Println("❌ [UPLOAD] Token Error:", err)
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: " + err.Error()})
		}

		// B. Validasi Achievement ID
		achID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			log.Println("❌ [UPLOAD] Invalid ID:", c.Params("id"))
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Achievement ID"})
		}

		// C. Ambil File
		file, err := c.FormFile("file")
		if err != nil {
			log.Println("❌ [UPLOAD] FormFile Error:", err)
			return c.Status(400).JSON(fiber.Map{
				"error":  "Failed to get file. Ensure key is 'file'",
				"detail": err.Error(),
			})
		}

		log.Printf("📂 [UPLOAD] File diterima: %s (%d bytes)\n", file.Filename, file.Size)

		// D. Simpan File
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		savePath := filepath.Join("./uploads", filename)

		if err := c.SaveFile(file, savePath); err != nil {
			log.Println("❌ [UPLOAD] SaveFile Error:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save file locally: " + err.Error()})
		}

		// E. Update Database
		fileURL := fmt.Sprintf("http://localhost:3000/uploads/%s", filename)
		dto := service.AttachmentDTO{
			FileName: file.Filename,
			FileURL:  fileURL,
			FileType: file.Header.Get("Content-Type"),
		}

		log.Println("🔄 [UPLOAD] Update DB Mongo...")
		if err := achievementService.UploadEvidence(c.Context(), userID, achID, dto); err != nil {
			log.Println("❌ [UPLOAD] Service Error:", err)
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		log.Println("✅ [UPLOAD] Sukses!")
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "File uploaded successfully",
			"data":    dto,
		})
	})
}
