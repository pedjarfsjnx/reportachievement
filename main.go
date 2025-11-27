package main

import (
	"context"
	"log"
	"os"

	// Import path menyesuaikan nama module Anda + folder database
	"reportachievement/database/mongo"
	"reportachievement/database/postgres"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 1. Init Database (Path Baru)
	dbPostgres := postgres.Connect()
	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	dbMongo := mongo.Connect()
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 2. Init Fiber
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Server berjalan dengan struktur folder baru!",
		})
	})

	log.Fatal(app.Listen(":" + os.Getenv("APP_PORT")))
}
