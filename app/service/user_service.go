package service

import (
	"errors"
	"reportachievement/app/model/postgre"
	repo "reportachievement/app/repository/postgre"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repo.UserRepository
}

func NewUserService(userRepo *repo.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// DTO: Input Create User
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	RoleName string `json:"role"` // "Admin", "Mahasiswa", "Dosen Wali"

	// Opsional: Data Profil
	NIM          string `json:"nim,omitempty"`           // Jika Mahasiswa
	ProgramStudy string `json:"program_study,omitempty"` // Jika Mahasiswa
	AcademicYear string `json:"academic_year,omitempty"` // Jika Mahasiswa
	LecturerID   string `json:"lecturer_id,omitempty"`   // Jika Dosen (NIP/NIDN)
	Department   string `json:"department,omitempty"`    // Jika Dosen
}

// DTO: Input Update User
type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	IsActive *bool  `json:"is_active"` // Pointer agar bisa detect false
}

// 1. Get All Users
func (s *UserService) GetAll() ([]postgre.User, error) {
	return s.userRepo.FindAll()
}

// 2. Create User (Complex Logic)
func (s *UserService) Create(req CreateUserRequest) error {
	// A. Cari Role ID berdasarkan nama role
	role, err := s.userRepo.FindRoleByName(req.RoleName)
	if err != nil {
		return errors.New("invalid role name: " + req.RoleName)
	}

	// B. Hash Password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// C. Siapkan Object User Utama
	newUser := postgre.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPwd),
		FullName:     req.FullName,
		RoleID:       role.ID,
		IsActive:     true,
	}

	// D. Tentukan apakah perlu buat profil tambahan
	if req.RoleName == "Mahasiswa" {
		if req.NIM == "" {
			return errors.New("NIM is required for Mahasiswa")
		}
		studentProfile := &postgre.Student{
			NIM:          req.NIM,
			ProgramStudy: req.ProgramStudy,
			AcademicYear: req.AcademicYear,
		}
		// Simpan User + Student dalam 1 transaksi
		return s.userRepo.CreateWithProfile(&newUser, studentProfile)

	} else if req.RoleName == "Dosen Wali" {
		if req.LecturerID == "" {
			return errors.New("lecturer_id (NIP) is required for Dosen Wali")
		}
		lecturerProfile := &postgre.Lecturer{
			LecturerID: req.LecturerID,
			Department: req.Department,
		}
		// Simpan User + Lecturer dalam 1 transaksi
		return s.userRepo.CreateWithProfile(&newUser, lecturerProfile)
	}

	// Jika Admin (tanpa profil khusus), simpan user biasa
	return s.userRepo.Create(&newUser)
}

// 3. Update User
func (s *UserService) Update(id uuid.UUID, req UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	return s.userRepo.Update(user)
}

// 4. Delete User
func (s *UserService) Delete(id uuid.UUID) error {
	// Cek dulu apakah user ada
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	return s.userRepo.Delete(id)
}
