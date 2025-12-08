package postgre

import (
	"reportachievement/app/service"
	"reportachievement/helper" // Import Helper

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Service *service.AuthService
}

type LoginRequest struct {
	Username string `json:"username" example:"superadmin"`
	Password string `json:"password" example:"admin123"`
}

func RegisterAuthRoutes(app *fiber.App, authService *service.AuthService) {
	h := &AuthHandler{Service: authService}
	api := app.Group("/api/v1/auth")
	api.Post("/login", h.Login)
}

// Swagger comments...
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	resp, err := h.Service.Login(req.Username, req.Password)
	if err != nil {
		return helper.Error(c, 401, err.Error())
	}

	return helper.Success(c, 200, "Login successful", resp)
}
