package postgre

import (
	"reportachievement/app/model/postgre"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LecturerRepository struct {
	db *gorm.DB
}

func NewLecturerRepository(db *gorm.DB) *LecturerRepository {
	return &LecturerRepository{db: db}
}

// FindByUserID: Mencari profil Dosen berdasarkan akun User ID yang login
func (r *LecturerRepository) FindByUserID(userID uuid.UUID) (*postgre.Lecturer, error) {
	var lecturer postgre.Lecturer
	// Cari di tabel lecturers dimana user_id cocok
	err := r.db.Where("user_id = ?", userID).First(&lecturer).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}
