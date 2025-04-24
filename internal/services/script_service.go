package services

import (
	"context"
	"errors"
	"fmt"

	"scripts-management/internal/models"
	"scripts-management/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScriptService struct {
	scriptRepo      *repository.ScriptRepository
	scriptShareRepo *repository.ScriptShareRepository
	userRepo        *repository.UserRepository
}

func NewScriptService(
	scriptRepo *repository.ScriptRepository,
	scriptShareRepo *repository.ScriptShareRepository,
	userRepo *repository.UserRepository,
) *ScriptService {
	return &ScriptService{
		scriptRepo:      scriptRepo,
		scriptShareRepo: scriptShareRepo,
		userRepo:        userRepo,
	}
}

func (s *ScriptService) CreateScript(ctx context.Context, userID primitive.ObjectID, req *models.CreateScriptRequest) (*models.Script, error) {
	script := &models.Script{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Type:        req.Type,
		OwnerID:     userID,
	}

	if err := s.scriptRepo.Create(ctx, script); err != nil {
		return nil, fmt.Errorf("failed to create script: %w", err)
	}

	return script, nil
}

func (s *ScriptService) GetScriptByID(ctx context.Context, userID, scriptID primitive.ObjectID) (*models.Script, error) {
	script, err := s.scriptRepo.FindByID(ctx, scriptID)
	if err != nil {
		return nil, fmt.Errorf("failed to find script: %w", err)
	}

	// Check if user is owner
	if script.OwnerID == userID {
		return script, nil
	}

	// Check if script is shared with user
	_, err = s.scriptShareRepo.FindByScriptIDAndUserID(ctx, scriptID, userID)
	if err != nil {
		return nil, errors.New("access denied: script not shared with user")
	}

	return script, nil
}

func (s *ScriptService) GetUserScripts(ctx context.Context, userID primitive.ObjectID) ([]*models.Script, error) {
	// Get scripts owned by user
	ownedScripts, err := s.scriptRepo.FindByOwnerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find owned scripts: %w", err)
	}

	// Get scripts shared with user
	shares, err := s.scriptShareRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shared scripts: %w", err)
	}

	// Get shared script details
	sharedScripts := make([]*models.Script, 0, len(shares))
	for _, share := range shares {
		script, err := s.scriptRepo.FindByID(ctx, share.ScriptID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				continue
			}
			return nil, fmt.Errorf("failed to find shared script: %w", err)
		}
		sharedScripts = append(sharedScripts, script)
	}

	// Combine owned and shared scripts
	allScripts := append(ownedScripts, sharedScripts...)
	return allScripts, nil
}

func (s *ScriptService) UpdateScript(ctx context.Context, userID, scriptID primitive.ObjectID, req *models.UpdateScriptRequest) (*models.Script, error) {
	script, err := s.scriptRepo.FindByID(ctx, scriptID)
	if err != nil {
		return nil, fmt.Errorf("failed to find script: %w", err)
	}

	// Only owner can update script
	if script.OwnerID != userID {
		return nil, errors.New("access denied: only owner can update script")
	}

	// Update fields if provided
	if req.Name != "" {
		script.Name = req.Name
	}
	if req.Description != "" {
		script.Description = req.Description
	}
	if req.Content != "" {
		script.Content = req.Content
	}
	if req.Type != "" {
		script.Type = req.Type
	}

	if err := s.scriptRepo.Update(ctx, script); err != nil {
		return nil, fmt.Errorf("failed to update script: %w", err)
	}

	return script, nil
}

func (s *ScriptService) DeleteScript(ctx context.Context, userID, scriptID primitive.ObjectID) error {
	script, err := s.scriptRepo.FindByID(ctx, scriptID)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	// Only owner can delete script
	if script.OwnerID != userID {
		return errors.New("access denied: only owner can delete script")
	}

	if err := s.scriptRepo.Delete(ctx, scriptID); err != nil {
		return fmt.Errorf("failed to delete script: %w", err)
	}

	return nil
}

func (s *ScriptService) ShareScript(ctx context.Context, ownerID, scriptID primitive.ObjectID, targetUserID string) error {
	script, err := s.scriptRepo.FindByID(ctx, scriptID)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	// Only owner can share script
	if script.OwnerID != ownerID {
		return errors.New("access denied: only owner can share script")
	}

	// Convert target user ID from string to ObjectID
	targetID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Check if target user exists
	_, err = s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		return errors.New("target user not found")
	}

	// Check if script is already shared with user
	_, err = s.scriptShareRepo.FindByScriptIDAndUserID(ctx, scriptID, targetID)
	if err == nil {
		return errors.New("script already shared with this user")
	}

	// Create share record
	share := &models.ScriptShare{
		ScriptID: scriptID,
		UserID:   targetID,
	}

	if err := s.scriptShareRepo.Create(ctx, share); err != nil {
		return fmt.Errorf("failed to share script: %w", err)
	}

	return nil
}

func (s *ScriptService) RevokeShare(ctx context.Context, ownerID, scriptID primitive.ObjectID, targetUserID string) error {
	script, err := s.scriptRepo.FindByID(ctx, scriptID)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	// Only owner can revoke share
	if script.OwnerID != ownerID {
		return errors.New("access denied: only owner can revoke share")
	}

	// Convert target user ID from string to ObjectID
	targetID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Delete share record
	if err := s.scriptShareRepo.Delete(ctx, scriptID, targetID); err != nil {
		return fmt.Errorf("failed to revoke share: %w", err)
	}

	return nil
}
