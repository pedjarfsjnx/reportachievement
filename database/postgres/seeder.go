package postgres

import (
	"log"
	"reportachievement/app/model/postgre"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) {
	// 1. Seed Roles
	roles := []postgre.Role{
		{Name: "Admin", Description: "Pengelola sistem full access"},
		{Name: "Mahasiswa", Description: "Pelapor prestasi"},
		{Name: "Dosen Wali", Description: "Verifikator prestasi"},
	}

	for _, role := range roles {
		if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
			log.Printf("❌ Gagal seed role %s: %v", role.Name, err)
		}
	}
	log.Println("✅ Roles seeded!")

	// 2. Seed Super Admin
	var adminRole postgre.Role
	db.Where("name = ?", "Admin").First(&adminRole)

	hashedPassAdmin, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := postgre.User{
		Username: "superadmin", Email: "admin@report.com", PasswordHash: string(hashedPassAdmin),
		FullName: "Super Admin", RoleID: adminRole.ID, IsActive: true,
	}
	if err := db.Where("username = ?", "superadmin").FirstOrCreate(&adminUser).Error; err != nil {
		log.Printf("❌ Gagal seed admin: %v", err)
	}

	// === TAMBAHAN MODUL 7: SEED MAHASISWA ===
	var mhsRole postgre.Role
	db.Where("name = ?", "Mahasiswa").First(&mhsRole)

	hashedPassMhs, _ := bcrypt.GenerateFromPassword([]byte("mhs123"), bcrypt.DefaultCost)

	// A. Buat User Mahasiswa
	mhsUser := postgre.User{
		Username: "mahasiswa1", Email: "mhs1@student.unair.ac.id", PasswordHash: string(hashedPassMhs),
		FullName: "Budi Santoso", RoleID: mhsRole.ID, IsActive: true,
	}

	// Simpan User dulu
	if err := db.Where("username = ?", "mahasiswa1").FirstOrCreate(&mhsUser).Error; err != nil {
		log.Printf("❌ Gagal seed user mahasiswa: %v", err)
	} else {
		// B. Buat Profile Student (Relasi ke User)
		studentProfile := postgre.Student{
			UserID:       mhsUser.ID,
			NIM:          "082011633001",
			ProgramStudy: "D4 Teknik Informatika",
			AcademicYear: "2024",
		}
		// FirstOrCreate berdasarkan NIM agar tidak duplikat
		if err := db.Where("nim = ?", "082011633001").FirstOrCreate(&studentProfile).Error; err != nil {
			log.Printf("❌ Gagal seed profile student: %v", err)
		} else {
			log.Println("✅ Dummy Mahasiswa created! (User: mahasiswa1 / Pass: mhs123)")
		}
	}
}
