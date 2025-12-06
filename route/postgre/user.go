package postgre

import (
	"reportachievement/app/service"
	"reportachievement/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	Service *service.UserService
}

func RegisterUserRoutes(app *fiber.App, userService *service.UserService) {
	h := &UserHandler{Service: userService}
	api := app.Group("/api/v1/users")
	api.Use(middleware.Protected())

	api.Get("/", h.GetAll)
	api.Post("/", h.Create)
	api.Put("/:id", h.Update)
	api.Delete("/:id", h.Delete)
}

func (h *UserHandler) isAdmin(c *fiber.Ctx) bool {
	return c.Locals("role") == "Admin"
}

// GetAll Users godoc
// @Summary      List Semua User (Admin)
// @Tags         User Management
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /users [get]
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}
	users, err := h.Service.GetAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "data": users})
}

// Create User godoc
// @Summary      Buat User Baru (Admin)
// @Tags         User Management
// @Param        request body service.CreateUserRequest true "User Data"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}
// @Router       /users [post]
func (h *UserHandler) Create(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}
	var req service.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if err := h.Service.Create(req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "User created"})
}

// Update User godoc
// @Summary      Update User (Admin)
// @Tags         User Management
// @Param        id   path      string  true  "User ID"
// @Param        request body service.UpdateUserRequest true "Update Data"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}
	id, _ := uuid.Parse(c.Params("id"))
	var req service.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	if err := h.Service.Update(id, req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Updated"})
}

// Delete User godoc
// @Summary      Hapus User (Admin)
// @Tags         User Management
// @Param        id   path      string  true  "User ID"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}
	id, _ := uuid.Parse(c.Params("id"))
	if err := h.Service.Delete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Deleted"})
}
