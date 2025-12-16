package repository

import (
	"context"
	mongoModel "reportachievement/app/model/mongo"
	"reportachievement/app/model/postgre"
	postgreRepo "reportachievement/app/repository/postgre" // Import untuk struct AchievementFilter

	"github.com/google/uuid"
)

// Interface untuk User Repository
type IUserRepository interface {
	FindByUsername(username string) (*postgre.User, error)
	FindByID(id uuid.UUID) (*postgre.User, error)
}

// Interface untuk Student Repository
type IStudentRepository interface {
	FindByUserID(userID uuid.UUID) (*postgre.Student, error)
	FindIDsByAdvisorID(advisorID uuid.UUID) ([]uuid.UUID, error)
}

// Interface untuk Lecturer Repository
type ILecturerRepository interface {
	FindByUserID(userID uuid.UUID) (*postgre.Lecturer, error)
}

// Interface untuk Achievement Repository (Postgres)
type IAchievementRepository interface {
	Create(data *postgre.AchievementReference) error
	FindAll(filter postgreRepo.AchievementFilter) ([]postgre.AchievementReference, int64, error)
	FindByID(id uuid.UUID) (*postgre.AchievementReference, error)
	VerifyOrReject(id uuid.UUID, updates map[string]interface{}) error
	UpdateStatus(id uuid.UUID, status string) error
}

// Interface untuk Achievement Repository (Mongo)
type IAchievementMongoRepository interface {
	Insert(ctx context.Context, data *mongoModel.Achievement) (string, error)
	FindByIDs(ctx context.Context, ids []string) ([]mongoModel.Achievement, error)
	SoftDelete(ctx context.Context, id string) error
	AddAttachment(ctx context.Context, achievementID string, attachment mongoModel.Attachment) error
}
