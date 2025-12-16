package mocks

import (
	"context"
	mongoModel "reportachievement/app/model/mongo"
	"reportachievement/app/model/postgre"
	postgreRepo "reportachievement/app/repository/postgre"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// --- MOCK USER REPO ---
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindByUsername(username string) (*postgre.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgre.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(id uuid.UUID) (*postgre.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgre.User), args.Error(1)
}

// --- MOCK STUDENT REPO ---
type MockStudentRepo struct {
	mock.Mock
}

func (m *MockStudentRepo) FindByUserID(userID uuid.UUID) (*postgre.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgre.Student), args.Error(1)
}

func (m *MockStudentRepo) FindIDsByAdvisorID(advisorID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(advisorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// --- MOCK LECTURER REPO ---
type MockLecturerRepo struct {
	mock.Mock
}

func (m *MockLecturerRepo) FindByUserID(userID uuid.UUID) (*postgre.Lecturer, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgre.Lecturer), args.Error(1)
}

// --- MOCK ACHIEVEMENT REPO ---
type MockAchRepo struct {
	mock.Mock
}

func (m *MockAchRepo) Create(data *postgre.AchievementReference) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockAchRepo) FindAll(filter postgreRepo.AchievementFilter) ([]postgre.AchievementReference, int64, error) {
	args := m.Called(filter)
	return args.Get(0).([]postgre.AchievementReference), args.Get(1).(int64), args.Error(2)
}

func (m *MockAchRepo) FindByID(id uuid.UUID) (*postgre.AchievementReference, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgre.AchievementReference), args.Error(1)
}

func (m *MockAchRepo) VerifyOrReject(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockAchRepo) UpdateStatus(id uuid.UUID, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

// --- MOCK MONGO REPO ---
type MockMongoRepo struct {
	mock.Mock
}

func (m *MockMongoRepo) Insert(ctx context.Context, data *mongoModel.Achievement) (string, error) {
	args := m.Called(ctx, data)
	return args.String(0), args.Error(1)
}

func (m *MockMongoRepo) FindByIDs(ctx context.Context, ids []string) ([]mongoModel.Achievement, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]mongoModel.Achievement), args.Error(1)
}

func (m *MockMongoRepo) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMongoRepo) AddAttachment(ctx context.Context, achievementID string, attachment mongoModel.Attachment) error {
	args := m.Called(ctx, achievementID, attachment)
	return args.Error(0)
}
