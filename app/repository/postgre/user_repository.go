package postgre

import (
	"reportachievement/app/model/postgre"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername mencari user dan meng-include data Role-nya
func (r *UserRepository) FindByUsername(username string) (*postgre.User, error) {
	var user postgre.User
	// Preload("Role") penting agar kita tahu role user tersebut (Admin/Mahasiswa)
	err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
