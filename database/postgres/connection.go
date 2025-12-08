package postgres

import (
	"log"
	"reportachievement/config" // Import Config

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	// Gunakan DSN dari Config
	dsn := cfg.PostgresDSN

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal koneksi ke PostgreSQL:", err)
	}

	log.Println("✅ Terkoneksi ke PostgreSQL!")
	return db
}
