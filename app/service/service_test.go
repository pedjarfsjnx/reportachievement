package service

import (
	"context"
	"log"
	"os"
	"testing"

	"reportachievement/app/model/postgre"
	repoMongo "reportachievement/app/repository/mongo"
	repoPostgre "reportachievement/app/repository/postgre"
	"reportachievement/config"
	"reportachievement/database/mongo"
	"reportachievement/database/postgres"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Global DB Connection untuk Testing
var (
	testDB       *gorm.DB
	testMongo    mongo.MongoInstance
	authService  *AuthService
	achService   *AchievementService
	userRepo     *repoPostgre.UserRepository
	studentRepo  *repoPostgre.StudentRepository
	lecturerRepo *repoPostgre.LecturerRepository
	achRefRepo   *repoPostgre.AchievementRepository
	achMongoRepo *repoMongo.AchievementRepository
)

// setup() berjalan sekali sebelum semua test dimulai
func TestMain(m *testing.M) {
	// 1. Load .env (Naik 2 folder ke root karena file ini ada di app/service)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("⚠️ Warning: .env not found, using system env")
	}

	// 2. Load Config & Connect DB
	cfg := config.LoadConfig()
	testDB = postgres.Connect(cfg)
	testMongo = mongo.Connect(cfg)

	// --- HARD RESET DATABASE (SOLUSI FINAL) ---
	// Kita gunakan Raw SQL 'CASCADE' untuk memaksa hapus tabel lama yang nyangkut.
	// Ini akan menghapus tabel bersih-bersih sebelum membuatnya lagi.
	testDB.Exec("DROP TABLE IF EXISTS achievement_references CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS students CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS lecturers CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS role_permissions CASCADE") // jika ada
	testDB.Exec("DROP TABLE IF EXISTS users CASCADE")
	testDB.Exec("DROP TABLE IF EXISTS roles CASCADE")

	// 3. Auto Migrate (Membuat tabel baru dengan struktur yang BENAR dari academic.go)
	err := testDB.AutoMigrate(
		&postgre.Role{},
		&postgre.User{},
		&postgre.Student{},
		&postgre.Lecturer{},
		&postgre.AchievementReference{},
	)
	if err != nil {
		log.Fatal("Gagal Migrasi Database Test:", err)
	}

	// 4. Setup Dependencies
	userRepo = repoPostgre.NewUserRepository(testDB)
	studentRepo = repoPostgre.NewStudentRepository(testDB)
	lecturerRepo = repoPostgre.NewLecturerRepository(testDB)
	achRefRepo = repoPostgre.NewAchievementRepository(testDB)
	achMongoRepo = repoMongo.NewAchievementRepository(testMongo.Db)

	authService = NewAuthService(userRepo)
	achService = NewAchievementService(studentRepo, lecturerRepo, achRefRepo, achMongoRepo)

	// 5. Jalankan Test
	code := m.Run()

	// 6. Cleanup Disconnect Mongo
	testMongo.Client.Disconnect(context.TODO())
	os.Exit(code)
}

// --- HELPER FUNCTION ---
func getOrCreateRole(name string) uuid.UUID {
	var role postgre.Role
	err := testDB.Where("name = ?", name).First(&role).Error
	if err == nil {
		return role.ID
	}
	newRole := postgre.Role{ID: uuid.New(), Name: name}
	testDB.Create(&newRole)
	return newRole.ID
}

// --- TEST 1: LOGIN (Auth Service) ---

func TestLogin_Integration(t *testing.T) {
	// A. Persiapan Data Dummy
	password := "rahasia123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	roleID := getOrCreateRole("Mahasiswa")

	dummyUser := postgre.User{
		ID:           uuid.New(),
		Username:     "test_login_user",
		Email:        "test@login.com",
		PasswordHash: string(hashed),
		FullName:     "Tester Login",
		RoleID:       roleID,
		IsActive:     true,
	}

	if err := testDB.Create(&dummyUser).Error; err != nil {
		t.Fatalf("Gagal insert user dummy: %v", err)
	}

	// CLEANUP SETELAH TEST
	defer testDB.Unscoped().Delete(&dummyUser)

	// B. Eksekusi Login (Happy Path)
	t.Run("Login Sukses", func(t *testing.T) {
		resp, err := authService.Login("test_login_user", password)

		assert.NoError(t, err)
		if err != nil {
			t.FailNow()
		}

		assert.NotNil(t, resp)

		// AKSES VIA MAP (Sesuai kode Auth Service Anda)
		assert.NotNil(t, resp["token"], "Token tidak boleh kosong")

		userData, ok := resp["user"].(map[string]interface{})
		assert.True(t, ok, "Data user harus berupa map")
		assert.Equal(t, "test_login_user", userData["username"])
	})

	// C. Eksekusi Login (Wrong Password)
	t.Run("Password Salah", func(t *testing.T) {
		_, err := authService.Login("test_login_user", "salah_pass")
		assert.Error(t, err)
		assert.Equal(t, "invalid username or password", err.Error())
	})
}

// --- TEST 2: ACHIEVEMENT FLOW (Create & Verify) ---

func TestAchievementFlow_Integration(t *testing.T) {
	// A. Persiapan Data
	passHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	roleMhsID := getOrCreateRole("Mahasiswa")
	roleDosenID := getOrCreateRole("Dosen Wali")

	// 1. Buat User Dosen
	dosenUser := postgre.User{
		ID: uuid.New(), Username: "dosen_test", Email: "d@test.com",
		PasswordHash: string(passHash), RoleID: roleDosenID, IsActive: true,
	}
	if err := testDB.Create(&dosenUser).Error; err != nil {
		t.Fatalf("Gagal buat Dosen: %v", err)
	}

	dosenProfile := postgre.Lecturer{
		ID: uuid.New(), UserID: dosenUser.ID,
		LecturerID: "NIP_TEST_001", // String ini sekarang valid karena struct sudah varchar
		Department: "IT",
	}
	if err := testDB.Create(&dosenProfile).Error; err != nil {
		t.Fatalf("Gagal buat Profil Dosen: %v", err)
	}

	// 2. Buat User Mahasiswa
	mhsUser := postgre.User{
		ID: uuid.New(), Username: "mhs_test", Email: "m@test.com",
		PasswordHash: string(passHash), RoleID: roleMhsID, IsActive: true,
	}
	if err := testDB.Create(&mhsUser).Error; err != nil {
		t.Fatalf("Gagal buat Mhs: %v", err)
	}

	mhsProfile := postgre.Student{
		ID: uuid.New(), UserID: mhsUser.ID,
		NIM:       "NIM_TEST_001", // String ini sekarang valid
		AdvisorID: &dosenProfile.ID,
	}
	if err := testDB.Create(&mhsProfile).Error; err != nil {
		t.Fatalf("Gagal buat Profil Mhs: %v", err)
	}

	// Variable ID Prestasi untuk cleanup
	var createdAchID uuid.UUID

	// --- DEFER CLEANUP ---
	defer func() {
		if createdAchID != uuid.Nil {
			testDB.Unscoped().Where("id = ?", createdAchID).Delete(&postgre.AchievementReference{})
		}
		testDB.Unscoped().Delete(&mhsProfile)
		testDB.Unscoped().Delete(&dosenProfile)
		testDB.Unscoped().Delete(&mhsUser)
		testDB.Unscoped().Delete(&dosenUser)
	}()

	// B. Test Case: Create Achievement
	t.Run("Create Achievement Draft", func(t *testing.T) {
		req := CreateAchievementRequest{
			Title: "Juara Lomba Coding", Type: "Kompetisi", Points: 100, Description: "Menang juara 1",
			Details: map[string]interface{}{"tingkat": "Nasional"},
		}

		res, err := achService.Create(context.Background(), mhsUser.ID, req)
		assert.NoError(t, err)
		if res == nil {
			t.Fatal("Response Nil")
		}

		createdAchID = res.ID
		assert.Equal(t, "draft", res.Status)
	})

	// C. Test Case: Submit
	t.Run("Submit Achievement", func(t *testing.T) {
		err := achService.Submit(context.Background(), mhsUser.ID, createdAchID)
		assert.NoError(t, err)

		var check postgre.AchievementReference
		testDB.First(&check, "id = ?", createdAchID)
		assert.Equal(t, "submitted", check.Status)
	})

	// D. Test Case: Dosen Verify
	t.Run("Dosen Verify Success", func(t *testing.T) {
		err := achService.Verify(context.Background(), dosenUser.ID, createdAchID)
		assert.NoError(t, err)

		var check postgre.AchievementReference
		testDB.First(&check, "id = ?", createdAchID)
		assert.Equal(t, "verified", check.Status)
		assert.Equal(t, dosenUser.ID, *check.VerifiedBy)
	})
}
