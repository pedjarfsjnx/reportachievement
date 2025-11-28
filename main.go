package main

import (
	"context"
	"log"
	"os"

	"reportachievement/app/model/postgre"
	"reportachievement/database/mongo"
	"reportachievement/database/postgres"

	repoPostgre "reportachievement/app/repository/postgre"
	"reportachievement/app/service"

	// Import route dari folder ROOT (Path baru)
	routePostgre "reportachievement/route/postgre"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Setup Env & DB
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	dbPostgres := postgres.Connect()

	// AutoMigrate & Seeding
	dbPostgres.AutoMigrate(&postgre.Role{}, &postgre.User{}, &postgre.Permission{}, &postgre.RolePermission{})
	postgres.SeedDatabase(dbPostgres)

	sqlDB, _ := dbPostgres.DB()
	defer sqlDB.Close()

	dbMongo := mongo.Connect()
	defer func() {
		if err := dbMongo.Client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	// 2. Setup Dependency Injection (Layer App)
	userRepo := repoPostgre.NewUserRepository(dbPostgres)
	authService := service.NewAuthService(userRepo)

	// 3. Setup Fiber
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	// 4. Register Routes (Layer Route/Delivery)
	routePostgre.RegisterAuthRoutes(app, authService)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Server Ready!"})
	})

	// 5. Run
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(app.Listen(port))
}
