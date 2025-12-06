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

// 1. FindByUserID
func (r *StudentRepository) FindByUserID(userID uuid.UUID) (*postgre.Student, error) {
	var student postgre.Student
	err := r.db.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// ---  ---

// 2. FindByIDs (Bulk Get untuk Mapping Nama di Leaderboard)
func (r *StudentRepository) FindByIDs(ids []uuid.UUID) ([]postgre.Student, error) {
	var students []postgre.Student
	// Preload User untuk dapat nama lengkap
	err := r.db.Preload("User").Where("id IN ?", ids).Find(&students).Error
	return students, err
}
