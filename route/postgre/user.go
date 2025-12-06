package postgre

import (
	"reportachievement/app/service"
	"reportachievement/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterUserRoutes(app *fiber.App, userService *service.UserService) {
	// Grouping route /api/v1/users
	api := app.Group("/api/v1/users")

	// Middleware: Protected (Harus Login)
	// Kita tambahkan pengecekan Role manual di dalam handler untuk memastikan hanya Admin
	api.Use(middleware.Protected())

	// Helper untuk cek Admin
	isAdmin := func(c *fiber.Ctx) bool {
		role := c.Locals("role")
		return role == "Admin"
	}

	// 1. GET ALL USERS
	api.Get("/", func(c *fiber.Ctx) error {
		if !isAdmin(c) {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden: Only Admin can access"})
		}

		users, err := userService.GetAll()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   users,
		})
	})

	// 2. CREATE USER
	api.Post("/", func(c *fiber.Ctx) error {
		if !isAdmin(c) {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
		}

		var req service.CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}

		if err := userService.Create(req); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(201).JSON(fiber.Map{
			"status":  "success",
			"message": "User created successfully",
		})
	})

	// 3. UPDATE USER
	api.Put("/:id", func(c *fiber.Ctx) error {
		if !isAdmin(c) {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
		}

		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
		}

		var req service.UpdateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}

		if err := userService.Update(id, req); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"status": "success", "message": "User updated"})
	})

	// 4. DELETE USER
	api.Delete("/:id", func(c *fiber.Ctx) error {
		if !isAdmin(c) {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
		}

		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
		}

		if err := userService.Delete(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"status": "success", "message": "User deleted"})
	})
}
