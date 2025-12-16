package service

import (
	"errors"
	"reportachievement/app/model/postgre"
	"reportachievement/app/repository" // Import Interface
	"reportachievement/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo repository.IUserRepository // Gunakan Interface
}

// Constructor terima Interface
func NewAuthService(userRepo repository.IUserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(username, password string) (map[string]interface{}, error) {
	// 1. Cari User
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 2. Cek Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	// 3. Generate JWT
	cfg := config.LoadConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role.Name,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role.Name,
			"fullName": user.FullName,
		},
	}, nil
}

// Get Profile ---
func (s *AuthService) GetProfile(userID uuid.UUID) (*postgre.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	//
	user.PasswordHash = ""
	return user, nil
}
