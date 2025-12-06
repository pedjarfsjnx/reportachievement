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

// Create
func (r *AchievementRepository) Create(data *postgre.AchievementReference) error {
	return r.db.Create(data).Error
}

// Filter Struct
type AchievementFilter struct {
	StudentID string
	Status    string
	Page      int
	Limit     int
}

// FindAll (List)
func (r *AchievementRepository) FindAll(filter AchievementFilter) ([]postgre.AchievementReference, int64, error) {
	var achievements []postgre.AchievementReference
	var total int64

	query := r.db.Model(&postgre.AchievementReference{})

	if filter.StudentID != "" {
		query = query.Where("student_id = ?", filter.StudentID)
	}

	// Jangan tampilkan status 'deleted' di list biasa
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		query = query.Where("status != ?", "deleted")
	}

	query.Count(&total)

	// Preload Relasi
	query = query.Preload("Student.User").Preload("Student").Preload("Verifier")

	offset := (filter.Page - 1) * filter.Limit
	err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&achievements).Error

	return achievements, total, err
}

// FindByID (Untuk detail & cek ownership)
func (r *AchievementRepository) FindByID(id uuid.UUID) (*postgre.AchievementReference, error) {
	var ach postgre.AchievementReference
	err := r.db.Preload("Student").First(&ach, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ach, nil
}

// UpdateStatus (Untuk mengubah jadi 'deleted')
func (r *AchievementRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&postgre.AchievementReference{}).Where("id = ?", id).Update("status", status).Error
}

// --- BARU (MODUL 10): Update Status Lengkap (Verifikasi/Reject/Submit) ---
func (r *AchievementRepository) VerifyOrReject(id uuid.UUID, data map[string]interface{}) error {
	// Updates memungkinkan kita update beberapa field sekaligus (status, verified_at, verified_by, rejection_note)
	return r.db.Model(&postgre.AchievementReference{}).Where("id = ?", id).Updates(data).Error
}
