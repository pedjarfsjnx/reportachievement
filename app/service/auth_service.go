package service

import (
	"errors"
	"os"
	"reportachievement/app/model/postgre"
	repo "reportachievement/app/repository/postgre"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repo.UserRepository
}

func NewAuthService(userRepo *repo.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Struct untuk response login
type LoginResponse struct {
	Token string        `json:"token"`
	User  *postgre.User `json:"user"`
}

func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	// 1. Cari user by username
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 2. Cek Password (Bcrypt)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 3. Generate JWT Token
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) generateJWT(user *postgre.User) (string, error) {
	// Secret key dari .env
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rahasia-default-jangan-dipakai-di-prod"
	}

	// Claims (isi payload token)
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role.Name,                        // Penting untuk RBAC nanti
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Expired 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
