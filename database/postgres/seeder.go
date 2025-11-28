package postgres

import (
	"log"
	"reportachievement/app/model/postgre"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) {
	// 1. Seed Roles (Sesuai SRS Hal 4)
	roles := []postgre.Role{
		{Name: "Admin", Description: "Pengelola sistem full access"},
		{Name: "Mahasiswa", Description: "Pelapor prestasi"},
		{Name: "Dosen Wali", Description: "Verifikator prestasi"},
	}

	for _, role := range roles {
		// FirstOrCreate: Buat data hanya jika belum ada (berdasarkan Name)
		if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
			log.Printf("❌ Gagal seed role %s: %v", role.Name, err)
		}
	}
	log.Println("✅ Roles seeded (Data Role Aman)!")

	// 2. Seed Super Admin User (Hanya jika belum ada)
	var adminRole postgre.Role
	// Cari Role ID untuk 'Admin' yang baru saja kita seed
	if err := db.Where("name = ?", "Admin").First(&adminRole).Error; err != nil {
		log.Printf("❌ Gagal mencari role Admin: %v", err)
		return
	}

	// Cek apakah user superadmin sudah ada?
	var count int64
	db.Model(&postgre.User{}).Where("username = ?", "superadmin").Count(&count)

	if count == 0 {
		// Hash Password
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

		adminUser := postgre.User{
			Username:     "superadmin",
			Email:        "admin@report.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Super Admin Sistem",
			RoleID:       adminRole.ID, // Link ke Role Admin
			IsActive:     true,
		}

		if err := db.Create(&adminUser).Error; err != nil {
			log.Printf("❌ Gagal create admin: %v", err)
		} else {
			log.Println("✅ Super Admin user created! (User: superadmin / Pass: admin123)")
		}
	} else {
		log.Println("ℹ️ Super Admin user sudah ada.")
	}
}
