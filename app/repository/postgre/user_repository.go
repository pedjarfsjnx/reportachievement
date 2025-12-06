package postgre

import (
	"reportachievement/app/model/postgre"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// 1. FindByUsername (Untuk Login - Sudah ada sebelumnya)
func (r *UserRepository) FindByUsername(username string) (*postgre.User, error) {
	var user postgre.User
	err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//  (CRUD USER) ---

// 2. FindAll (List Users)
func (r *UserRepository) FindAll() ([]postgre.User, error) {
	var users []postgre.User
	// Preload Role agar admin tau user ini role-nya apa
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

// 3. FindByID
func (r *UserRepository) FindByID(id uuid.UUID) (*postgre.User, error) {
	var user postgre.User
	err := r.db.Preload("Role").First(&user, "id = ?", id).Error
	return &user, err
}

// 4. Create User (Standard)
func (r *UserRepository) Create(user *postgre.User) error {
	return r.db.Create(user).Error
}

// 5. Create User dengan Transaction (Untuk User + Student/Lecturer Profile)
// Fungsi ini menerima callback function di dalamnya
func (r *UserRepository) CreateWithProfile(user *postgre.User, profile interface{}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create User dulu
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 2. Create Profile (Student/Lecturer) sesuai input
		// Kita perlu assign UserID dari user yang baru dibuat ke profile
		switch p := profile.(type) {
		case *postgre.Student:
			p.UserID = user.ID
			if err := tx.Create(p).Error; err != nil {
				return err
			}
		case *postgre.Lecturer:
			p.UserID = user.ID
			if err := tx.Create(p).Error; err != nil {
				return err
			}
		}

		return nil // Commit transaction
	})
}

// 6. Update
func (r *UserRepository) Update(user *postgre.User) error {
	return r.db.Save(user).Error
}

// 7. Delete (Hard Delete atau Soft Delete via IsActive)
// Di sini kita pakai Hard Delete data user, gorm akan handle cascade jika disetting.
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&postgre.User{}, "id = ?", id).Error
}

// Helper: Find Role by Name (Agar kita bisa cari ID role berdasarkan string "Mahasiswa")
func (r *UserRepository) FindRoleByName(name string) (*postgre.Role, error) {
	var role postgre.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}
