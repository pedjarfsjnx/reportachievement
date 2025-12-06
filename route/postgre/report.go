package postgre

import (
	"reportachievement/app/service"
	"reportachievement/middleware"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	Service *service.ReportService
}

func RegisterReportRoutes(app *fiber.App, reportService *service.ReportService) {
	h := &ReportHandler{Service: reportService}
	api := app.Group("/api/v1/reports")
	api.Use(middleware.Protected())

	api.Get("/statistics", h.GetStats)
}

// GetStats godoc
// @Summary      Dashboard Statistik
// @Description  Melihat ranking dan statistik prestasi
// @Tags         Reporting
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /reports/statistics [get]
func (h *ReportHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.Service.GetDashboardStats(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "data": stats})
}
