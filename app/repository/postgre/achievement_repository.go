package postgre

import (
	"reportachievement/app/model/postgre"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementRepository struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

// UPDATE STRUCT FILTER
type AchievementFilter struct {
	Page       int
	Limit      int
	Status     string
	StudentIDs []uuid.UUID // <-- TAMBAHAN: Filter Array ID Mahasiswa
}

// 1. CREATE
func (r *AchievementRepository) Create(data *postgre.AchievementReference) error {
	return r.db.Create(data).Error
}

// 2. FIND ALL (UPDATE QUERY)
func (r *AchievementRepository) FindAll(filter AchievementFilter) ([]postgre.AchievementReference, int64, error) {
	var achievements []postgre.AchievementReference
	var total int64

	// Base Query (Preload User & Student data)
	query := r.db.Model(&postgre.AchievementReference{}).
		Preload("Student").
		Preload("Student.User")

	// Filter Status
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		// Default: Jangan tampilkan yang deleted
		query = query.Where("status != ?", "deleted")
	}

	// --- Filter Spesifik Mahasiswa ---
	if len(filter.StudentIDs) > 0 {
		query = query.Where("student_id IN ?", filter.StudentIDs)
	}
	// ----------------------------------------------

	// Hitung Total Data (Untuk Pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (filter.Page - 1) * filter.Limit
	err := query.Limit(filter.Limit).Offset(offset).Order("created_at DESC").Find(&achievements).Error

	return achievements, total, err
}

// 3. FIND BY ID
func (r *AchievementRepository) FindByID(id uuid.UUID) (*postgre.AchievementReference, error) {
	var achievement postgre.AchievementReference
	err := r.db.Preload("Student").
		Preload("Student.User").
		First(&achievement, "id = ?", id).Error
	return &achievement, err
}

// 4. VERIFY OR REJECT (Update Status)
func (r *AchievementRepository) VerifyOrReject(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&postgre.AchievementReference{}).Where("id = ?", id).Updates(updates).Error
}

// 5. UPDATE STATUS (Soft Delete)
func (r *AchievementRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&postgre.AchievementReference{}).Where("id = ?", id).Update("status", status).Error
}
