package postgre

import (
	"reportachievement/app/model/postgre"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Cari data student berdasarkan user_id (dari token JWT nanti)
func (r *StudentRepository) FindByUserID(userID uuid.UUID) (*postgre.Student, error) {
	var student postgre.Student
	err := r.db.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}
