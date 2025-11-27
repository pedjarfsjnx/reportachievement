package postgre

import (
	"time"

	"github.com/google/uuid"
)

// SRS Halaman 4: 3.1.2 Tabel roles [cite: 40]
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(50);unique;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SRS Halaman 4: 3.1.1 Tabel users [cite: 36]
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username     string    `gorm:"type:varchar(50);unique;not null"`
	Email        string    `gorm:"type:varchar(100);unique;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	FullName     string    `gorm:"type:varchar(100);not null"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	Role         Role      `gorm:"foreignKey:RoleID"` // Relasi ke tabel Role
	IsActive     bool      `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SRS Halaman 4: 3.1.3 Permissions & 3.1.4 RolePermissions [cite: 51, 67]
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(100);unique;not null"` // e.g., achievement:create
	Resource    string    `gorm:"type:varchar(50);not null"`
	Action      string    `gorm:"type:varchar(50);not null"`
	Description string    `gorm:"type:text"`
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"primaryKey;type:uuid"`
	PermissionID uuid.UUID `gorm:"primaryKey;type:uuid"`
}
