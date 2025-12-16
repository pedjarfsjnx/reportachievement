package postgre

import (
	"time"

	"github.com/google/uuid"
)

// Tabel lecturers
type Lecturer struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	// Ganti column:lecturer_id menjadi column:nip
	// JSON tetap "lecturer_id" sesuai SRS
	LecturerID string `gorm:"type:varchar(20);unique;not null;column:nip" json:"lecturer_id"`

	Department string    `gorm:"type:varchar(100)" json:"department"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Tabel students
type Student struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	// Ganti column:student_id menjadi column:nim
	// Agar GORM tidak mengira ini Foreign Key ke achievement
	// JSON tetap "student_id" sesuai SRS
	NIM string `gorm:"type:varchar(20);unique;not null;column:nim" json:"student_id"`

	ProgramStudy string     `gorm:"type:varchar(100)" json:"program_study"`
	AcademicYear string     `gorm:"type:varchar(10)" json:"academic_year"`
	AdvisorID    *uuid.UUID `gorm:"type:uuid" json:"advisor_id"`
	Advisor      *Lecturer  `gorm:"foreignKey:AdvisorID" json:"advisor"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
