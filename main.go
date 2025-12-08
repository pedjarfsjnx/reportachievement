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

	_ "reportachievement/docs"

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

	// 1. Init DB Postgres
	dbPostgres := postgres.Connect()
	dbPostgres.AutoMigrate(
		&postgre.Role{}, &postgre.User{}, &postgre.Permission{}, &postgre.RolePermission{},
		&postgre.Lecturer{}, &postgre.Student{}, &postgre.AchievementReference{},
	)
	postgres.SeedDatabase(dbPostgres)
	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	// 2. Init DB Mongo
	dbMongo := mongo.Connect()
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 3. Dependency Injection

	// -- REPOSITORIES --
	userRepo := repoPostgre.NewUserRepository(dbPostgres)
	studentRepo := repoPostgre.NewStudentRepository(dbPostgres)
	achRefRepo := repoPostgre.NewAchievementRepository(dbPostgres)
	lecturerRepo := repoPostgre.NewLecturerRepository(dbPostgres) // <-- 1. REPO BARU
	achMongoRepo := repoMongo.NewAchievementRepository(dbMongo.Db)

	// -- SERVICES --
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)

	// <-- 2. INJECT LECTURER REPO KE ACHIEVEMENT SERVICE
	achService := service.NewAchievementService(studentRepo, lecturerRepo, achRefRepo, achMongoRepo)

	reportService := service.NewReportService(achMongoRepo, studentRepo)

	// 4. Init Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New())

	// 5. Static Files
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
	app.Static("/uploads", "./uploads")
	app.Static("/", "./public")

	// 6. Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 7. Register Routes
	routePostgre.RegisterAuthRoutes(app, authService)
	routePostgre.RegisterAchievementRoutes(app, achService)
	routePostgre.RegisterUserRoutes(app, userService)
	routePostgre.RegisterReportRoutes(app, reportService)

	// 8. Run Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(app.Listen(port))
}
