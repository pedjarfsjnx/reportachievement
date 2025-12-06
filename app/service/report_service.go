package service

import (
	"context"
	mongoRepo "reportachievement/app/repository/mongo"
	postgreRepo "reportachievement/app/repository/postgre"

	"github.com/google/uuid"
)

type ReportService struct {
	mongoRepo   *mongoRepo.AchievementRepository
	studentRepo *postgreRepo.StudentRepository
}

func NewReportService(mongoRepo *mongoRepo.AchievementRepository, studentRepo *postgreRepo.StudentRepository) *ReportService {
	return &ReportService{
		mongoRepo:   mongoRepo,
		studentRepo: studentRepo,
	}
}

// Struct Response Dashboard
type DashboardStats struct {
	TopStudents        []TopStudentDTO `json:"top_students"`
	AchievementsByType map[string]int  `json:"achievements_by_type"`
}

type TopStudentDTO struct {
	Rank        int    `json:"rank"`
	Name        string `json:"name"`
	NIM         string `json:"nim"`
	TotalPoints int    `json:"total_points"`
	TotalCount  int    `json:"total_achievements"`
}

func (s *ReportService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	// 1. Ambil Statistik per Tipe dari Mongo
	typeStats, err := s.mongoRepo.GetStatsByType(ctx)
	if err != nil {
		return nil, err
	}

	// Convert ke Map biar JSON rapi
	typeMap := make(map[string]int)
	for _, t := range typeStats {
		typeMap[t.Key] = t.Count
	}

	// 2. Ambil Top 5 Mahasiswa dari Mongo (berdasarkan Poin)
	topList, err := s.mongoRepo.GetTopStudents(ctx, 5)
	if err != nil {
		return nil, err
	}

	// 3. Ambil Nama Mahasiswa dari Postgres
	var studentIDs []uuid.UUID
	for _, item := range topList {
		if id, err := uuid.Parse(item.StudentPostgresID); err == nil {
			studentIDs = append(studentIDs, id)
		}
	}

	students, err := s.studentRepo.FindByIDs(studentIDs)
	if err != nil {
		return nil, err
	}

	// Mapping ID -> Student Struct
	studentMap := make(map[string]string) // ID -> Name
	nimMap := make(map[string]string)     // ID -> NIM
	for _, stu := range students {
		studentMap[stu.ID.String()] = stu.User.FullName
		nimMap[stu.ID.String()] = stu.NIM
	}

	// 4. Gabungkan Data (Ranking)
	var rankList []TopStudentDTO
	for i, item := range topList {
		name := studentMap[item.StudentPostgresID]
		if name == "" {
			name = "Unknown Student"
		}

		rankList = append(rankList, TopStudentDTO{
			Rank:        i + 1,
			Name:        name,
			NIM:         nimMap[item.StudentPostgresID],
			TotalPoints: item.TotalPoints,
			TotalCount:  item.TotalAchievements,
		})
	}

	return &DashboardStats{
		TopStudents:        rankList,
		AchievementsByType: typeMap,
	}, nil
}
