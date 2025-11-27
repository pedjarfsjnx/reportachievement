package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 2. Init App
	app := fiber.New()

	// 3. Middlewares Dasar
	app.Use(cors.New())
	app.Use(logger.New())

	// 4. Test Route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Server Report Achievement berjalan!",
			"status":  "success",
		})
	})

	// 5. Listen
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Fatal(app.Listen(port))
}
