package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// Import Models
	mongoModel "reportachievement/app/model/mongo"
	postgreModel "reportachievement/app/model/postgre"

	// Import Repositories
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

// Request Body untuk Create Achievement
type CreateAchievementRequest struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // competition, organization, etc
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"` // Data dinamis
	Points      int                    `json:"points"`
}

// Response Body untuk List Achievement (Gabungan Postgres & Mongo)
type AchievementListResponse struct {
	ID          uuid.UUID `json:"id"` // ID Postgres (Reference ID)
	Status      string    `json:"status"`
	StudentName string    `json:"student_name"`
	NIM         string    `json:"nim"`

	// Data dari MongoDB
	Title   string                 `json:"title"`
	Type    string                 `json:"type"`
	Points  int                    `json:"points"`
	Details map[string]interface{} `json:"details"`

	CreatedAt string `json:"created_at"`
}

// --- METHODS ---

// 1. CREATE ACHIEVEMENT (Modul 7)
func (s *AchievementService) Create(ctx context.Context, userID uuid.UUID, req CreateAchievementRequest) (*postgreModel.AchievementReference, error) {
	// A. Cari Student Profile berdasarkan User ID
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// B. Siapkan Data MongoDB
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

	// C. Simpan ke MongoDB -> Dapat ID String
	mongoID, err := s.achMongoRepo.Insert(ctx, mongoData)
	if err != nil {
		return nil, errors.New("failed to save to mongodb: " + err.Error())
	}

	// D. Siapkan Data PostgreSQL
	pgData := &postgreModel.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// E. Simpan ke PostgreSQL
	if err := s.achRefRepo.Create(pgData); err != nil {
		return nil, errors.New("failed to save reference: " + err.Error())
	}

	return pgData, nil
}

// 2. GET ALL ACHIEVEMENTS (Modul 8)
func (s *AchievementService) GetAll(ctx context.Context, filter postgreRepo.AchievementFilter) ([]AchievementListResponse, int64, error) {
	// A. Ambil Data Referensi dari PostgreSQL (dengan Filter & Pagination)
	pgData, total, err := s.achRefRepo.FindAll(filter)
	if err != nil {
		return nil, 0, err
	}

	// Jika data kosong, langsung return array kosong
	if len(pgData) == 0 {
		return []AchievementListResponse{}, 0, nil
	}

	// B. Kumpulkan semua MongoID dari hasil query Postgres
	var mongoIDs []string
	for _, item := range pgData {
		mongoIDs = append(mongoIDs, item.MongoAchievementID)
	}

	// C. Ambil Detail dari MongoDB berdasarkan list ID tadi (Bulk Query)
	mongoDocs, err := s.achMongoRepo.FindByIDs(ctx, mongoIDs)
	if err != nil {
		return nil, 0, err
	}

	// D. Mapping Data Mongo ke Map agar mudah diakses berdasarkan ID
	mongoMap := make(map[string]mongoModel.Achievement)
	for _, doc := range mongoDocs {
		mongoMap[doc.ID.Hex()] = doc
	}

	// E. Gabungkan (Merge) Data Postgres dan Mongo
	var response []AchievementListResponse
	for _, pg := range pgData {
		// Cari detail di map
		mongoDetail, exists := mongoMap[pg.MongoAchievementID]

		res := AchievementListResponse{
			ID:        pg.ID,
			Status:    pg.Status,
			CreatedAt: pg.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// Ambil data user/student dari preload Postgres
		if pg.Student.User.FullName != "" {
			res.StudentName = pg.Student.User.FullName
			res.NIM = pg.Student.NIM
		} else {
			res.StudentName = "Unknown"
			res.NIM = "-"
		}

		// Masukkan data Mongo jika ada
		if exists {
			res.Title = mongoDetail.Title
			res.Type = mongoDetail.AchievementType
			res.Points = mongoDetail.Points
			res.Details = mongoDetail.Details
		} else {
			res.Title = "[Data Missing in Mongo]"
		}

		response = append(response, res)
	}

	return response, total, nil
}
