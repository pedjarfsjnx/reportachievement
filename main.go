package main

import (
	"context"
	"log"
	"os"

	"reportachievement/database/mongo"
	"reportachievement/database/postgres"

	// Models (Domain Layer)
	"reportachievement/app/model/postgre"

	// Repositories (Data Layer)
	repoMongo "reportachievement/app/repository/mongo"
	repoPostgre "reportachievement/app/repository/postgre"

	// Services (Business Logic Layer)
	"reportachievement/app/service"

	// Routes (Transport Layer - DI LUAR APP)
	routePostgre "reportachievement/route/postgre"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 2. Init DB Postgres
	dbPostgres := postgres.Connect()

	// AutoMigrate
	log.Println("⏳ Menjalankan Migrasi Database...")
	err := dbPostgres.AutoMigrate(
		&postgre.Role{}, &postgre.User{}, &postgre.Permission{}, &postgre.RolePermission{},
		&postgre.Lecturer{}, &postgre.Student{}, &postgre.AchievementReference{},
	)
	if err != nil {
		log.Fatal("❌ Gagal Migrasi: ", err)
	}

	// Seeding
	postgres.SeedDatabase(dbPostgres)

	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	// 3. Init DB Mongo
	dbMongo := mongo.Connect()
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 4. Setup Dependency Injection
	// --- AUTH Module ---
	userRepo := repoPostgre.NewUserRepository(dbPostgres)
	authService := service.NewAuthService(userRepo)

	// --- ACHIEVEMENT Module ---
	studentRepo := repoPostgre.NewStudentRepository(dbPostgres)
	achRefRepo := repoPostgre.NewAchievementRepository(dbPostgres)
	achMongoRepo := repoMongo.NewAchievementRepository(dbMongo.Db)

	achService := service.NewAchievementService(studentRepo, achRefRepo, achMongoRepo)

	// 5. Setup Fiber
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	// 6. Register Routes
	routePostgre.RegisterAuthRoutes(app, authService)
	routePostgre.RegisterAchievementRoutes(app, achService)

	// Default Route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Server Report Achievement berjalan!"})
	})

	// 7. Run Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(app.Listen(port))
}
