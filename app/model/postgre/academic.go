package postgre

import (
	"time"

	"github.com/google/uuid"
)

// SRS Halaman 5: 3.1.6 Tabel lecturers
type Lecturer struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	User       User      `gorm:"foreignKey:UserID"`
	LecturerID string    `gorm:"type:varchar(20);unique;not null"` // NIP/NIDN
	Department string    `gorm:"type:varchar(100)"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SRS Halaman 5: 3.1.5 Tabel students
type Student struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null"`
	User   User      `gorm:"foreignKey:UserID"`

	// --- PERBAIKAN DI SINI ---
	// Ganti "StudentID" menjadi "NIM" agar GORM tidak bingung dengan Foreign Key
	NIM string `gorm:"type:varchar(20);unique;not null"`
	// -------------------------

	ProgramStudy string     `gorm:"type:varchar(100)"`
	AcademicYear string     `gorm:"type:varchar(10)"` // e.g., "2024/2025"
	AdvisorID    *uuid.UUID `gorm:"type:uuid"`        // Boleh null jika belum dapat dosen wali
	Advisor      *Lecturer  `gorm:"foreignKey:AdvisorID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
