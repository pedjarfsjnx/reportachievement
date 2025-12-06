package postgre

import (
	"reportachievement/app/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Service *service.AuthService
}

// Struct Request Body (Exported agar terbaca Swagger)
type LoginRequest struct {
	Username string `json:"username" example:"superadmin"`
	Password string `json:"password" example:"admin123"`
}

func RegisterAuthRoutes(app *fiber.App, authService *service.AuthService) {
	h := &AuthHandler{Service: authService}
	api := app.Group("/api/v1/auth")

	api.Post("/login", h.Login)
}

// Login godoc
// @Summary      Masuk ke sistem
// @Description  Autentikasi user dan mendapatkan JWT Token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Credentials"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid request body"})
	}

	resp, err := h.Service.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Login successful",
		"data":    resp,
	})
}
