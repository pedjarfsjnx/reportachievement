package postgre

import (
	"fmt"
	"reportachievement/app/service"
	"reportachievement/helper"
	"reportachievement/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	// --- TAMBAHAN BARU ---
	api.Get("/profile", middleware.Protected(), h.GetProfile) // Butuh Token
	api.Post("/logout", h.Logout)                             // Logout (Stateless)
}

// Login godoc
// @Summary      Login User
// @Description  Authenticate user and get JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Credentials"
// @Success      200  {object} helper.APIResponse
// @Failure      400  {object} helper.APIResponse
// @Router       /api/v1/auth/login [post]
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

// GetProfile godoc
// @Summary      Get User Profile
// @Description  Get currently logged in user profile
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} helper.APIResponse
// @Failure      401  {object} helper.APIResponse
// @Router       /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Ambil ID dari Token (set via Middleware)
	claims := c.Locals("user_id")
	if claims == nil {
		return helper.Error(c, 401, "Unauthorized")
	}

	// Konversi interface{} ke UUID
	userID, err := uuid.Parse(fmt.Sprintf("%v", claims))
	if err != nil {
		return helper.Error(c, 400, "Invalid User ID Token")
	}

	user, err := h.Service.GetProfile(userID)
	if err != nil {
		return helper.Error(c, 404, "User not found")
	}

	return helper.Success(c, 200, "User Profile", user)
}

// Logout godoc
// @Summary      Logout User
// @Description  Logout user (client side clear token)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object} helper.APIResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Karena pakai JWT (Stateless), server hanya perlu kirim respon OK.
	// Client yang bertugas menghapus token dari local storage.
	return helper.Success(c, 200, "Successfully logged out", nil)
}
