package main

import (
	"context"
	"log"
	"os"

	// Import package database lokal
	"reportachievement/database/mongo"
	"reportachievement/database/postgres"

	// Import model untuk kebutuhan migrasi (Modul 3)
	"reportachievement/app/model/postgre"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 2. Inisialisasi Database PostgreSQL
	dbPostgres := postgres.Connect()

	// === AUTO MIGRATION (MODUL 3) ===
	log.Println("⏳ Menjalankan Migrasi Database...")
	err := dbPostgres.AutoMigrate(
		&postgre.Role{},
		&postgre.User{},
		&postgre.Permission{},
		&postgre.RolePermission{},
	)
	if err != nil {
		log.Fatal("❌ Gagal Migrasi: ", err)
	}
	log.Println("✅ Migrasi Berhasil!")
	// ================================

	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	// 3. Inisialisasi Database MongoDB
	dbMongo := mongo.Connect()
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 4. Inisialisasi Fiber App
	app := fiber.New()

	// 5. Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// 6. Test Route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":  "Server Report Achievement berjalan!",
			"database": "PostgreSQL & MongoDB Connected",
			"status":   "success",
		})
	})

	// 7. Jalankan Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(app.Listen(port))
}
