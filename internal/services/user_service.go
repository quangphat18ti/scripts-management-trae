package services

import (
	"context"
	"errors"
	"fmt"

	"scripts-management/internal/config"
	"scripts-management/internal/models"
	"scripts-management/internal/repository"
	"scripts-management/pkg/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo    *repository.UserRepository
	authService *AuthService
	config      *config.Config
}

func NewUserService(userRepo *repository.UserRepository, config *config.Config, auth *AuthService) (*UserService, error) {
	if userRepo == nil {
		return nil, errors.New("userRepo cannot be nil")
	}
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if auth == nil {
		return nil, errors.New("auth service cannot be nil")
	}

	return &UserService{
		userRepo:    userRepo,
		config:      config,
		authService: auth,
	}, nil
}

func (s *UserService) InitRootAccount(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	existing, err := s.userRepo.FindByUsername(ctx, s.config.RootUsername)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("failed to check existing root account: %w", err)
	}

	if existing != nil {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.config.RootPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash root password: %w", err)
	}

	root := &models.User{
		Username: s.config.RootUsername,
		Password: string(hashedPassword),
		Role:     models.RoleRoot,
	}

	if err := s.userRepo.Create(ctx, root); err != nil {
		return fmt.Errorf("failed to create root account: %w", err)
	}

	return nil
}

func (s *UserService) CreateUser(ctx context.Context, currentUser *utils.JWTClaims, newUser *models.SignupRequest) error {
	currentRole := models.UserRole(currentUser.Role)

	if currentRole != models.RoleRoot && currentRole != models.RoleAdmin {
		return errors.New("insufficient permissions")
	}

	if currentRole == models.RoleAdmin && newUser.Role == models.RoleAdmin {
		return errors.New("admin cannot create other admin accounts")
	}

	return s.authService.CreateUser(ctx, newUser)
}

func (s *UserService) DeleteUser(ctx context.Context, currentUser *utils.JWTClaims, userID primitive.ObjectID) error {
	targetUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	currentRole := models.UserRole(currentUser.Role)

	if currentRole == models.RoleAdmin {
		if targetUser.Role == models.RoleAdmin || targetUser.Role == models.RoleRoot {
			return errors.New("insufficient permissions")
		}
	}

	return s.userRepo.Delete(ctx, userID)
}

func (s *UserService) ChangePassword(ctx context.Context, currentUser *utils.JWTClaims, userID primitive.ObjectID, newPassword string) error {
	targetUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	currentRole := models.UserRole(currentUser.Role)

	if currentRole == models.RoleAdmin {
		if targetUser.Role == models.RoleAdmin || targetUser.Role == models.RoleRoot {
			return errors.New("insufficient permissions")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}
