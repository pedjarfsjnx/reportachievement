package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SRS Halaman 6: 3.2.1 Collection achievements
type Achievement struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentPostgresID string             `bson:"student_postgres_id" json:"student_postgres_id"` // Referensi balik ke Postgres (UUID string)

	AchievementType string `bson:"achievement_type" json:"achievement_type"` // academic, competition, organization, etc.
	Title           string `bson:"title" json:"title"`
	Description     string `bson:"description" json:"description"`

	// Field dinamis (Competition details, Publication details, dll) disimpan dalam Map
	Details map[string]interface{} `bson:"details" json:"details"`

	Attachments []Attachment `bson:"attachments" json:"attachments"`
	Tags        []string     `bson:"tags" json:"tags"`
	Points      int          `bson:"points" json:"points"`

	// Untuk Soft Delete di Mongo (FR-005)
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Sub-struct untuk Attachments
type Attachment struct {
	FileName   string    `bson:"file_name" json:"file_name"`
	FileURL    string    `bson:"file_url" json:"file_url"`
	FileType   string    `bson:"file_type" json:"file_type"`
	UploadedAt time.Time `bson:"uploaded_at" json:"uploaded_at"`
}
