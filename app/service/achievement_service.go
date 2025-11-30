package service

import (
	"context"
	"errors"
	mongoModel "reportachievement/app/model/mongo"
	postgreModel "reportachievement/app/model/postgre"
	mongoRepo "reportachievement/app/repository/mongo"
	postgreRepo "reportachievement/app/repository/postgre"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService struct {
	studentRepo  *postgreRepo.StudentRepository
	achRefRepo   *postgreRepo.AchievementRepository
	achMongoRepo *mongoRepo.AchievementRepository
}

func NewAchievementService(
	studentRepo *postgreRepo.StudentRepository,
	achRefRepo *postgreRepo.AchievementRepository,
	achMongoRepo *mongoRepo.AchievementRepository,
) *AchievementService {
	return &AchievementService{
		studentRepo:  studentRepo,
		achRefRepo:   achRefRepo,
		achMongoRepo: achMongoRepo,
	}
}

// Struct untuk Request Input (DTO)
type CreateAchievementRequest struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // competition, organization, etc
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"` // Data dinamis (juara, tingkat, dll)
	Points      int                    `json:"points"`
}

func (s *AchievementService) Create(ctx context.Context, userID uuid.UUID, req CreateAchievementRequest) (*postgreModel.AchievementReference, error) {
	// 1. Cari Student Profile berdasarkan User ID yang login
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// 2. Siapkan Data MongoDB
	mongoData := &mongoModel.Achievement{
		ID:                primitive.NewObjectID(),
		StudentPostgresID: student.ID.String(),
		AchievementType:   req.Type,
		Title:             req.Title,
		Description:       req.Description,
		Details:           req.Details,
		Points:            req.Points,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// 3. Simpan ke MongoDB
	mongoID, err := s.achMongoRepo.Insert(ctx, mongoData)
	if err != nil {
		return nil, errors.New("failed to save to mongodb: " + err.Error())
	}

	// 4. Siapkan Data PostgreSQL (Reference)
	pgData := &postgreModel.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             "draft", // Default status
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 5. Simpan ke PostgreSQL
	if err := s.achRefRepo.Create(pgData); err != nil {
		// Note: Idealnya jika ini gagal, kita hapus data di Mongo (Rollback/Saga pattern).
		// Tapi untuk tahap ini, kita return error dulu.
		return nil, errors.New("failed to save reference: " + err.Error())
	}

	return pgData, nil
}
