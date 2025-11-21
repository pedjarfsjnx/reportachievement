package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Inisialisasi aplikasi Fiber
	app := fiber.New()

	// Tentukan rute (route) sederhana
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Halo, World! This is Go Fiber!")
	})

	// Jalankan server di port 3000
	log.Fatal(app.Listen(":3000"))
}
