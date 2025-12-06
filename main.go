package main

import (
	"context"
	"log"
	"os"

	"reportachievement/database/mongo"
	"reportachievement/database/postgres"

	"reportachievement/app/model/postgre"

	repoMongo "reportachievement/app/repository/mongo"
	repoPostgre "reportachievement/app/repository/postgre"

	"reportachievement/app/service"

	routePostgre "reportachievement/route/postgre"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 1. Init DB
	dbPostgres := postgres.Connect()
	dbPostgres.AutoMigrate(
		&postgre.Role{}, &postgre.User{}, &postgre.Permission{}, &postgre.RolePermission{},
		&postgre.Lecturer{}, &postgre.Student{}, &postgre.AchievementReference{},
	)
	postgres.SeedDatabase(dbPostgres)
	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	dbMongo := mongo.Connect()
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 2. Dependency Injection
	userRepo := repoPostgre.NewUserRepository(dbPostgres)
	authService := service.NewAuthService(userRepo)

	studentRepo := repoPostgre.NewStudentRepository(dbPostgres)
	achRefRepo := repoPostgre.NewAchievementRepository(dbPostgres)
	achMongoRepo := repoMongo.NewAchievementRepository(dbMongo.Db)

	achService := service.NewAchievementService(studentRepo, achRefRepo, achMongoRepo)

	// 3. Init Fiber
	app := fiber.New(fiber.Config{
		// Naikkan limit jadi 50MB agar aman
		BodyLimit: 50 * 1024 * 1024,
	})

	// --- MIDDLEWARE ---
	app.Use(logger.New())  // Log dulu biar kelihatan request masuk
	app.Use(recover.New()) // Tangkap panic
	app.Use(cors.New())    // Handle CORS

	// 4. Folder Upload
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		log.Println("📂 Membuat folder uploads...")
		os.Mkdir("./uploads", 0755)
	}
	app.Static("/uploads", "./uploads")

	// 5. Routes
	routePostgre.RegisterAuthRoutes(app, authService)
	routePostgre.RegisterAchievementRoutes(app, achService)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Server OK"})
	})

	// 6. Run
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(app.Listen(port))
}
