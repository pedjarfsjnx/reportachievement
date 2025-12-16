package postgre

import (
	"time"

	"github.com/google/uuid"
)

//	Tabel achievement_references
//
// Ditambah status 'deleted'
type AchievementReference struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StudentID uuid.UUID `gorm:"type:uuid;not null"`
	Student   Student   `gorm:"foreignKey:StudentID"`

	MongoAchievementID string `gorm:"type:varchar(24);not null"` // ID dari MongoDB

	// Enum Status: draft, submitted, verified, rejected, deleted
	Status string `gorm:"type:varchar(20);default:'draft'"`

	SubmittedAt *time.Time
	VerifiedAt  *time.Time
	VerifiedBy  *uuid.UUID `gorm:"type:uuid"` // Relasi ke User (Dosen/Admin)
	Verifier    *User      `gorm:"foreignKey:VerifiedBy"`

	RejectionNote string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
