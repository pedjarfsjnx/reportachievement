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

// 1. CREATE ACHIEVEMENT
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

// 2. GET ALL ACHIEVEMENTS
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
			res.Title = "[Deleted or Missing]"
		}

		response = append(response, res)
	}

	return response, total, nil
}

// 3. DELETE (Soft Delete)
func (s *AchievementService) Delete(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error {
	ach, err := s.achRefRepo.FindByID(achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("student profile not found")
	}

	if ach.StudentID != student.ID {
		return errors.New("unauthorized: you do not own this achievement")
	}

	if ach.Status != "draft" {
		return errors.New("cannot delete achievement with status: " + ach.Status)
	}

	if err := s.achMongoRepo.SoftDelete(ctx, ach.MongoAchievementID); err != nil {
		return errors.New("failed to delete mongo data: " + err.Error())
	}

	if err := s.achRefRepo.UpdateStatus(ach.ID, "deleted"); err != nil {
		return errors.New("failed to update status: " + err.Error())
	}

	return nil
}

// --- BARU (MODUL 10): SUBMIT, VERIFY, REJECT ---

// 4. SUBMIT (Mahasiswa)
func (s *AchievementService) Submit(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error {
	ach, err := s.achRefRepo.FindByID(achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return errors.New("student profile not found")
	}
	if ach.StudentID != student.ID {
		return errors.New("unauthorized action")
	}

	if ach.Status != "draft" {
		return errors.New("only draft achievement can be submitted")
	}

	now := time.Now()
	updateData := map[string]interface{}{
		"status":       "submitted",
		"submitted_at": &now,
	}

	return s.achRefRepo.VerifyOrReject(ach.ID, updateData)
}

// 5. VERIFY (Dosen Wali)
func (s *AchievementService) Verify(ctx context.Context, lecturerUserID uuid.UUID, achievementID uuid.UUID) error {
	ach, err := s.achRefRepo.FindByID(achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	if ach.Status != "submitted" {
		return errors.New("achievement is not in submitted status")
	}

	now := time.Now()
	updateData := map[string]interface{}{
		"status":      "verified",
		"verified_at": &now,
		"verified_by": lecturerUserID,
	}

	return s.achRefRepo.VerifyOrReject(ach.ID, updateData)
}

// 6. REJECT (Dosen Wali)
func (s *AchievementService) Reject(ctx context.Context, lecturerUserID uuid.UUID, achievementID uuid.UUID, note string) error {
	ach, err := s.achRefRepo.FindByID(achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	if ach.Status != "submitted" {
		return errors.New("achievement is not in submitted status")
	}

	now := time.Now()
	updateData := map[string]interface{}{
		"status":         "rejected",
		"verified_at":    &now,
		"verified_by":    lecturerUserID,
		"rejection_note": note,
	}

	return s.achRefRepo.VerifyOrReject(ach.ID, updateData)
}
