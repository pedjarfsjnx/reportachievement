package postgre

import (
	"reportachievement/app/service"
	"reportachievement/helper" // Import Helper
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

func (h *ReportHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.Service.GetDashboardStats(c.Context())
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, 200, "Dashboard Statistics", stats)
}
