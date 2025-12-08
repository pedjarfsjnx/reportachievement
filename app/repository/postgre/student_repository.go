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

// 1. FindByUserID (Untuk create prestasi)
func (r *StudentRepository) FindByUserID(userID uuid.UUID) (*postgre.Student, error) {
	var student postgre.Student
	err := r.db.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// 2. FindByIDs (Untuk Report Service)
func (r *StudentRepository) FindByIDs(ids []uuid.UUID) ([]postgre.Student, error) {
	var students []postgre.Student
	err := r.db.Preload("User").Where("id IN ?", ids).Find(&students).Error
	return students, err
}

// --- Cari List ID Mahasiswa Bimbingan Dosen Tertentu ---
func (r *StudentRepository) FindIDsByAdvisorID(advisorID uuid.UUID) ([]uuid.UUID, error) {
	var students []postgre.Student
	// Ambil hanya kolom ID biar cepat
	err := r.db.Select("id").Where("advisor_id = ?", advisorID).Find(&students).Error
	if err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	for _, s := range students {
		ids = append(ids, s.ID)
	}
	return ids, nil
}
