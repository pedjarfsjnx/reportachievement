package postgre

import (
	"reportachievement/app/model/postgre"

	"gorm.io/gorm"
)

type AchievementRepository struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func (r *AchievementRepository) Create(data *postgre.AchievementReference) error {
	return r.db.Create(data).Error
}

// TAMBAHAN MODUL 8: Filter Parameter
type AchievementFilter struct {
	StudentID string
	Status    string
	Page      int
	Limit     int
}

func (r *AchievementRepository) FindAll(filter AchievementFilter) ([]postgre.AchievementReference, int64, error) {
	var achievements []postgre.AchievementReference
	var total int64

	query := r.db.Model(&postgre.AchievementReference{})

	// Apply Filters
	if filter.StudentID != "" {
		query = query.Where("student_id = ?", filter.StudentID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Hitung total data (untuk pagination)
	query.Count(&total)

	// Preload Relasi: Student -> User, Verifier -> User
	query = query.Preload("Student.User").Preload("Student").Preload("Verifier")

	// Apply Pagination
	offset := (filter.Page - 1) * filter.Limit
	err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&achievements).Error

	return achievements, total, err
}
