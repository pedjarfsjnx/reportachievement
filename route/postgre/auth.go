package postgre

import (
	"reportachievement/app/service"

	"github.com/gofiber/fiber/v2"
)

// Request Body Struct (Didefinisikan di sini karena hanya dipakai di route ini)
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Fungsi ini menerima App (Fiber) dan Service yang dibutuhkan
func RegisterAuthRoutes(app *fiber.App, authService *service.AuthService) {

	api := app.Group("/api/v1/auth")

	// POST /api/v1/auth/login
	api.Post("/login", func(c *fiber.Ctx) error {
		// 1. Parsing Request Body
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid request body",
			})
		}

		// 2. Panggil Service Login
		resp, err := authService.Login(req.Username, req.Password)
		if err != nil {
			// Jika gagal login (password salah / user tak ada)
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		// 3. Return Success JSON
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Login successful",
			"data":    resp,
		})
	})
}
