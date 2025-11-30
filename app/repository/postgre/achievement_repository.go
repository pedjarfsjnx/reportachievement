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
