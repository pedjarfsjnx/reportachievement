package postgre

import (
	"fmt"
	"path/filepath"
	"reportachievement/app/repository/postgre"
	"reportachievement/app/service"
	"reportachievement/helper" // Asumsi pakai helper yang sudah dibuat
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

func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals("user_id")
	if claims == nil {
		return uuid.Nil, fmt.Errorf("user_id not found")
	}
	return uuid.Parse(fmt.Sprintf("%v", claims))
}

func (h *AchievementHandler) Create(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, err.Error())
	}
	var req service.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, 400, "Invalid JSON")
	}
	result, err := h.Service.Create(c.Context(), userID, req)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, 201, "Achievement draft created", result)
}

// GET LIST (UPDATE UTAMA)
func (h *AchievementHandler) GetList(c *fiber.Ctx) error {
	// 1. Ambil User ID dari Token
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, "Unauthorized")
	}

	// 2. Ambil Role dari Token (Diset di Middleware)
	role := c.Locals("role").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status")

	filter := postgre.AchievementFilter{Page: page, Limit: limit, Status: status}

	// 3. Panggil Service dengan UserID dan Role
	data, total, err := h.Service.GetAll(c.Context(), userID, role, filter)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, 200, "Success", fiber.Map{
		"data": data,
		"meta": fiber.Map{"page": page, "limit": limit, "total": total},
	})
}

// DELETE

func (h *AchievementHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, "Unauthorized")
	}
	achID, _ := uuid.Parse(c.Params("id"))
	if err := h.Service.Delete(c.Context(), userID, achID); err != nil {
		return helper.Error(c, 400, err.Error())
	}
	return helper.Success(c, 200, "Deleted", nil)
}

func (h *AchievementHandler) Submit(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, "Unauthorized")
	}
	achID, _ := uuid.Parse(c.Params("id"))
	if err := h.Service.Submit(c.Context(), userID, achID); err != nil {
		return helper.Error(c, 400, err.Error())
	}
	return helper.Success(c, 200, "Submitted", nil)
}

func (h *AchievementHandler) Verify(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, "Unauthorized")
	}
	achID, _ := uuid.Parse(c.Params("id"))
	if err := h.Service.Verify(c.Context(), userID, achID); err != nil {
		return helper.Error(c, 400, err.Error())
	}
	return helper.Success(c, 200, "Verified", nil)
}

func (h *AchievementHandler) Reject(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, "Unauthorized")
	}
	achID, _ := uuid.Parse(c.Params("id"))
	var req struct {
		Note string `json:"note"`
	}
	c.BodyParser(&req)
	if err := h.Service.Reject(c.Context(), userID, achID, req.Note); err != nil {
		return helper.Error(c, 400, err.Error())
	}
	return helper.Success(c, 200, "Rejected", nil)
}

func (h *AchievementHandler) UploadEvidence(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return helper.Error(c, 401, err.Error())
	}
	achID, _ := uuid.Parse(c.Params("id"))
	file, err := c.FormFile("file")
	if err != nil {
		return helper.Error(c, 400, "File required")
	}
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	savePath := filepath.Join("./uploads", filename)
	if err := c.SaveFile(file, savePath); err != nil {
		return helper.Error(c, 500, "Save failed")
	}
	fileURL := fmt.Sprintf("http://localhost:3000/uploads/%s", filename)
	dto := service.AttachmentDTO{FileName: file.Filename, FileURL: fileURL, FileType: file.Header.Get("Content-Type")}
	if err := h.Service.UploadEvidence(c.Context(), userID, achID, dto); err != nil {
		return helper.Error(c, 400, err.Error())
	}
	return helper.Success(c, 200, "Upload Success", dto)
}
