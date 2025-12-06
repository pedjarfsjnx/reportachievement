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

	// Import Swagger
	_ "reportachievement/docs" // Folder ini akan dibuat otomatis oleh 'swag init'

	"github.com/gofiber/swagger"
)

// @title           Sistem Pelaporan Prestasi Mahasiswa API
// @version         1.0
// @description     Dokumentasi API untuk Project UAS Backend Lanjut.
// @contact.name    Tim Pengembang
// @host            localhost:3000
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	// 2. DI (Dependency Injection)
	userRepo := repoPostgre.NewUserRepository(dbPostgres)
	studentRepo := repoPostgre.NewStudentRepository(dbPostgres)
	achRefRepo := repoPostgre.NewAchievementRepository(dbPostgres)
	achMongoRepo := repoMongo.NewAchievementRepository(dbMongo.Db)

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	achService := service.NewAchievementService(studentRepo, achRefRepo, achMongoRepo)
	reportService := service.NewReportService(achMongoRepo, studentRepo)

	// 3. Init Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New())

	// 4. Static Files
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
	app.Static("/uploads", "./uploads")

	// 5. Swagger Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 6. Register API Routes
	routePostgre.RegisterAuthRoutes(app, authService)
	routePostgre.RegisterAchievementRoutes(app, achService)
	routePostgre.RegisterUserRoutes(app, userService)
	routePostgre.RegisterReportRoutes(app, reportService)

	// 7. Run
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(app.Listen(port))
}
