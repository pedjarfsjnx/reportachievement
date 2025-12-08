package postgres

import (
	"fmt"
	"log"
	"reportachievement/app/model/postgre"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) {
	// 1. Cek apakah database sudah ada isinya?
	var count int64
	db.Model(&postgre.User{}).Count(&count)
	if count > 0 {
		log.Println("⚠️  Database sudah berisi data. Seeder dilewati.")
		return
	}

	log.Println("🌱 Memulai Seeding Data (15 User)...")

	// 2. CREATE ROLES
	roleAdmin := postgre.Role{ID: uuid.New(), Name: "Admin", Description: "Administrator"}
	roleDosen := postgre.Role{ID: uuid.New(), Name: "Dosen Wali", Description: "Verifikator"}
	roleMhs := postgre.Role{ID: uuid.New(), Name: "Mahasiswa", Description: "Pelapor"}

	roles := []postgre.Role{roleAdmin, roleDosen, roleMhs}
	if err := db.Create(&roles).Error; err != nil {
		log.Fatal("❌ Gagal seed roles:", err)
	}

	// Password Hash "123456"
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	strPwd := string(hashedPwd)

	// 3. CREATE SUPERADMIN
	admin := postgre.User{
		ID:           uuid.New(),
		Username:     "superadmin",
		Email:        "admin@unair.ac.id",
		PasswordHash: strPwd,
		FullName:     "Super Administrator",
		RoleID:       roleAdmin.ID,
		IsActive:     true,
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Fatal("Gagal buat admin:", err)
	}

	// 4. CREATE 5 DOSEN WALI
	var lecturerProfileIDs []uuid.UUID

	for i := 1; i <= 5; i++ {
		// A. User Dosen
		userID := uuid.New()
		user := postgre.User{
			ID:           userID,
			Username:     fmt.Sprintf("dosen%d", i),
			Email:        fmt.Sprintf("dosen%d@unair.ac.id", i),
			PasswordHash: strPwd,
			FullName:     fmt.Sprintf("Dr. Dosen %d", i),
			RoleID:       roleDosen.ID,
			IsActive:     true,
		}
		db.Create(&user)

		// B. Profil Dosen
		lecturerID := uuid.New()
		lecturer := postgre.Lecturer{
			ID:         lecturerID,
			UserID:     userID,
			LecturerID: fmt.Sprintf("NIP%03d", i),
			Department: "Teknik Informatika",
		}
		db.Create(&lecturer)

		lecturerProfileIDs = append(lecturerProfileIDs, lecturerID)
		log.Printf("✅ Dosen %d Created (User: dosen%d)", i, i)
	}

	// 5. CREATE 10 MAHASISWA
	for i := 1; i <= 10; i++ {
		// A. User Mahasiswa
		userID := uuid.New()
		user := postgre.User{
			ID:           userID,
			Username:     fmt.Sprintf("mhs%d", i),
			Email:        fmt.Sprintf("mhs%d@unair.ac.id", i),
			PasswordHash: strPwd,
			FullName:     fmt.Sprintf("Mahasiswa %d", i),
			RoleID:       roleMhs.ID,
			IsActive:     true,
		}
		db.Create(&user)

		// B. Tentukan Dosen Wali (Logic: 2 Mahasiswa per Dosen)
		lecturerIndex := (i - 1) / 2
		assignedAdvisorID := lecturerProfileIDs[lecturerIndex]

		// C. Profil Mahasiswa (FIXED: Menggunakan Field NIM)
		student := postgre.Student{
			ID:           uuid.New(),
			UserID:       userID,
			NIM:          fmt.Sprintf("NIM%03d", i), // <--- UBAH 'StudentID' JADI 'NIM'
			ProgramStudy: "Sistem Informasi",
			AcademicYear: "2024",
			AdvisorID:    &assignedAdvisorID,
		}
		db.Create(&student)

		log.Printf("✅ Mahasiswa %d Created -> Wali: Dosen %d", i, lecturerIndex+1)
	}

	log.Println("🎉 Seeding Selesai! Login Password: '123456'")
}
