package postgre

import (
	"reportachievement/app/service"
	"reportachievement/helper" // Import Helper
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

func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return helper.Error(c, 403, "Forbidden")
	}
	users, err := h.Service.GetAll()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, 200, "List Users", users)
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return helper.Error(c, 403, "Forbidden")
	}
	var req service.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, 400, "Invalid JSON")
	}
	if err := h.Service.Create(req); err != nil {
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, 201, "User created", nil)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return helper.Error(c, 403, "Forbidden")
	}
	id, _ := uuid.Parse(c.Params("id"))
	var req service.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, 400, "Invalid JSON")
	}
	if err := h.Service.Update(id, req); err != nil {
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, 200, "User Updated", nil)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return helper.Error(c, 403, "Forbidden")
	}
	id, _ := uuid.Parse(c.Params("id"))
	if err := h.Service.Delete(id); err != nil {
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, 200, "User Deleted", nil)
}
