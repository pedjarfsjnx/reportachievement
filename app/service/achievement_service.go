package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mongoModel "reportachievement/app/model/mongo"
	postgreModel "reportachievement/app/model/postgre"
	mongoRepo "reportachievement/app/repository/mongo"
	postgreRepo "reportachievement/app/repository/postgre"
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

// --- STRUCTS (DTO) ---
type CreateAchievementRequest struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Points      int                    `json:"points"`
}

type AchievementListResponse struct {
	ID          uuid.UUID              `json:"id"`
	Status      string                 `json:"status"`
	StudentName string                 `json:"student_name"`
	NIM         string                 `json:"nim"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Points      int                    `json:"points"`
	Details     map[string]interface{} `json:"details"`
	CreatedAt   string                 `json:"created_at"`
}

// --- METHODS ---

// 1. Create (Modul 7)
func (s *AchievementService) Create(ctx context.Context, userID uuid.UUID, req CreateAchievementRequest) (*postgreModel.AchievementReference, error) {
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

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

	mongoID, err := s.achMongoRepo.Insert(ctx, mongoData)
	if err != nil {
		return nil, errors.New("failed to save to mongodb: " + err.Error())
	}

	pgData := &postgreModel.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.achRefRepo.Create(pgData); err != nil {
		return nil, errors.New("failed to save reference: " + err.Error())
	}

	return pgData, nil
}

// 2. Get All (Modul 8)
func (s *AchievementService) GetAll(ctx context.Context, filter postgreRepo.AchievementFilter) ([]AchievementListResponse, int64, error) {
	pgData, total, err := s.achRefRepo.FindAll(filter)
	if err != nil {
		return nil, 0, err
	}

	if len(pgData) == 0 {
		return []AchievementListResponse{}, 0, nil
	}

	var mongoIDs []string
	for _, item := range pgData {
		mongoIDs = append(mongoIDs, item.MongoAchievementID)
	}

	mongoDocs, err := s.achMongoRepo.FindByIDs(ctx, mongoIDs)
	if err != nil {
		return nil, 0, err
	}

	mongoMap := make(map[string]mongoModel.Achievement)
	for _, doc := range mongoDocs {
		mongoMap[doc.ID.Hex()] = doc
	}

	var response []AchievementListResponse
	for _, pg := range pgData {
		mongoDetail, exists := mongoMap[pg.MongoAchievementID]

		res := AchievementListResponse{
			ID:        pg.ID,
			Status:    pg.Status,
			CreatedAt: pg.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if pg.Student.User.FullName != "" {
			res.StudentName = pg.Student.User.FullName
			res.NIM = pg.Student.NIM
		} else {
			res.StudentName = "Unknown"
			res.NIM = "-"
		}

		if exists {
			res.Title = mongoDetail.Title
			res.Type = mongoDetail.AchievementType
			res.Points = mongoDetail.Points
			res.Details = mongoDetail.Details
		} else {
			// Data mongo mungkin kena soft delete atau hilang
			res.Title = "[Deleted or Missing]"
		}

		response = append(response, res)
	}

	return response, total, nil
}

// 3. Delete (BARU - MODUL 9)
func (s *AchievementService) Delete(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error {
	// A. Cari Data Prestasi
	ach, err := s.achRefRepo.FindByID(achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	// B. Cek Ownership (Validasi apakah Student yg login adalah pemilik prestasi)
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("student profile not found")
	}

	if ach.StudentID != student.ID {
		return errors.New("unauthorized: you do not own this achievement")
	}

	// C. Cek Status (Hanya boleh hapus DRAFT)
	if ach.Status != "draft" {
		return errors.New("cannot delete achievement with status: " + ach.Status)
	}

	// D. Soft Delete di MongoDB
	if err := s.achMongoRepo.SoftDelete(ctx, ach.MongoAchievementID); err != nil {
		return errors.New("failed to delete mongo data: " + err.Error())
	}

	// E. Update Status di PostgreSQL jadi 'deleted'
	if err := s.achRefRepo.UpdateStatus(ach.ID, "deleted"); err != nil {
		return errors.New("failed to update status: " + err.Error())
	}

	return nil
}
