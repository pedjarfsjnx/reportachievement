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

type AchievementHandler struct {
	Service *service.AchievementService
}

func RegisterAchievementRoutes(app *fiber.App, achievementService *service.AchievementService) {
	h := &AchievementHandler{Service: achievementService}
	api := app.Group("/api/v1/achievements")

	api.Post("/", middleware.Protected(), h.Create)
	api.Get("/", middleware.Protected(), h.GetList)
	api.Delete("/:id", middleware.Protected(), h.Delete)
	api.Post("/:id/submit", middleware.Protected(), h.Submit)
	api.Post("/:id/verify", middleware.Protected(), h.Verify)
	api.Post("/:id/reject", middleware.Protected(), h.Reject)
	api.Post("/:id/attachments", middleware.Protected(), h.UploadEvidence)
}

// Helper: Get User ID from Token
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals("user_id")
	if claims == nil {
		return uuid.Nil, fmt.Errorf("user_id not found")
	}
	return uuid.Parse(fmt.Sprintf("%v", claims))
}

// Create godoc
// @Summary      Buat Draft Prestasi
// @Description  Mahasiswa membuat draft prestasi baru
// @Tags         Achievement
// @Accept       json
// @Produce      json
// @Param        request body service.CreateAchievementRequest true "Data Prestasi"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}
// @Router       /achievements [post]
func (h *AchievementHandler) Create(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	var req service.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	result, err := h.Service.Create(c.Context(), userID, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "data": result})
}

// GetList godoc
// @Summary      Lihat Daftar Prestasi
// @Description  Melihat list prestasi dengan pagination dan filter
// @Tags         Achievement
// @Produce      json
// @Param        page   query    int     false  "Page" default(1)
// @Param        limit  query    int     false  "Limit" default(10)
// @Param        status query    string  false  "Status Filter"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /achievements [get]
func (h *AchievementHandler) GetList(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status")

	filter := postgre.AchievementFilter{Page: page, Limit: limit, Status: status}
	data, total, err := h.Service.GetAll(c.Context(), filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   data,
		"meta":   fiber.Map{"page": page, "limit": limit, "total": total},
	})
}

// Delete godoc
// @Summary      Hapus Prestasi (Soft Delete)
// @Tags         Achievement
// @Param        id   path      string  true  "Achievement ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /achievements/{id} [delete]
func (h *AchievementHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	achID, _ := uuid.Parse(c.Params("id"))

	if err := h.Service.Delete(c.Context(), userID, achID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Deleted"})
}

// Submit godoc
// @Summary      Submit Prestasi
// @Description  Mengubah status draft menjadi submitted
// @Tags         Workflow
// @Param        id   path      string  true  "Achievement ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /achievements/{id}/submit [post]
func (h *AchievementHandler) Submit(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	achID, _ := uuid.Parse(c.Params("id"))

	if err := h.Service.Submit(c.Context(), userID, achID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Submitted"})
}

// Verify godoc
// @Summary      Verifikasi Prestasi (Dosen)
// @Tags         Workflow
// @Param        id   path      string  true  "Achievement ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /achievements/{id}/verify [post]
func (h *AchievementHandler) Verify(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	achID, _ := uuid.Parse(c.Params("id"))

	if err := h.Service.Verify(c.Context(), userID, achID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Verified"})
}

// Reject godoc
// @Summary      Tolak Prestasi (Dosen)
// @Tags         Workflow
// @Param        id   path      string  true  "Achievement ID"
// @Param        request body map[string]string true "Note Penolakan"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /achievements/{id}/reject [post]
func (h *AchievementHandler) Reject(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	achID, _ := uuid.Parse(c.Params("id"))

	var req struct {
		Note string `json:"note"`
	}
	c.BodyParser(&req)

	if err := h.Service.Reject(c.Context(), userID, achID, req.Note); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Rejected"})
}

// UploadEvidence godoc
// @Summary      Upload Bukti Prestasi
// @Tags         Achievement
// @Accept       mpfd
// @Param        id   path      string  true  "Achievement ID"
// @Param        file formData  file    true  "File Bukti"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /achievements/{id}/attachments [post]
func (h *AchievementHandler) UploadEvidence(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}
	achID, _ := uuid.Parse(c.Params("id"))

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File required"})
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	savePath := filepath.Join("./uploads", filename)
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Save failed"})
	}

	fileURL := fmt.Sprintf("http://localhost:3000/uploads/%s", filename)
	dto := service.AttachmentDTO{
		FileName: file.Filename, FileURL: fileURL, FileType: file.Header.Get("Content-Type"),
	}

	if err := h.Service.UploadEvidence(c.Context(), userID, achID, dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("✅ [UPLOAD] Success:", filename)
	return c.JSON(fiber.Map{"status": "success", "data": dto})
}
