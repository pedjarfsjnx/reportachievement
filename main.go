package main

import (
	"context"
	"log"
	"os"

	"reportachievement/config" // Import Config
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

	_ "reportachievement/docs"

	"github.com/gofiber/swagger"
)

// Swagger annotations...
func main() {
	// 1. Load Config (Menggantikan godotenv.Load manual)
	cfg := config.LoadConfig()

	// 2. Setup File Logging (Menulis ke folder /logs)
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", 0755)
	}
	file, err := os.OpenFile("./logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer file.Close()
	log.SetOutput(file) // Mengarahkan log.Println ke file

	// 3. Init DB (Pass Config ke fungsi koneksi)
	dbPostgres := postgres.Connect(cfg)
	dbPostgres.AutoMigrate(
		&postgre.Role{}, &postgre.User{}, &postgre.Permission{}, &postgre.RolePermission{},
		&postgre.Lecturer{}, &postgre.Student{}, &postgre.AchievementReference{},
	)

	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	dbMongo := mongo.Connect(cfg)
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 4. Dependency Injection
	userRepo := repoPostgre.NewUserRepository(dbPostgres)
	studentRepo := repoPostgre.NewStudentRepository(dbPostgres)
	achRefRepo := repoPostgre.NewAchievementRepository(dbPostgres)
	lecturerRepo := repoPostgre.NewLecturerRepository(dbPostgres)
	achMongoRepo := repoMongo.NewAchievementRepository(dbMongo.Db)

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	achService := service.NewAchievementService(studentRepo, lecturerRepo, achRefRepo, achMongoRepo)
	reportService := service.NewReportService(achMongoRepo, studentRepo)

	// 5. Init Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(cors.New())
	// Logger Middleware menulis ke file juga
	app.Use(logger.New(logger.Config{
		Output: file,
	}))

	// 6. Static Files
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}
	app.Static("/uploads", "./uploads")
	app.Static("/", "./public")

	// 7. Routes
	app.Get("/swagger/*", swagger.HandlerDefault)
	routePostgre.RegisterAuthRoutes(app, authService)
	routePostgre.RegisterAchievementRoutes(app, achService)
	routePostgre.RegisterUserRoutes(app, userService)
	routePostgre.RegisterReportRoutes(app, reportService)

	// 8. Run
	log.Println("🚀 Server running on port", cfg.AppPort)
	log.Fatal(app.Listen(cfg.AppPort))
}
