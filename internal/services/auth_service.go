package services

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"scripts-management/internal/models"
	"scripts-management/internal/repository"
	"scripts-management/pkg/utils"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *utils.JWTManager
}

func NewAuthService(userRepo *repository.UserRepository, jwtManager *utils.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (string, error) {
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.jwtManager.GenerateToken(user.ID, user.Username, string(user.Role))
}

func (s *AuthService) Signup(ctx context.Context, req *models.SignupRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	return s.userRepo.Create(ctx, user)
}