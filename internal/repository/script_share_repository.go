package repository

import (
	"context"
	"time"

	"scripts-management/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScriptShareRepository struct {
	collection *mongo.Collection
}

func NewScriptShareRepository(db *mongo.Database) *ScriptShareRepository {
	return &ScriptShareRepository{
		collection: db.Collection("script_shares"),
	}
}

func (r *ScriptShareRepository) Create(ctx context.Context, share *models.ScriptShare) error {
	share.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, share)
	return err
}

func (r *ScriptShareRepository) FindByScriptIDAndUserID(ctx context.Context, scriptID, userID primitive.ObjectID) (*models.ScriptShare, error) {
	var share models.ScriptShare
	err := r.collection.FindOne(ctx, bson.M{
		"script_id": scriptID,
		"user_id":   userID,
	}).Decode(&share)
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *ScriptShareRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.ScriptShare, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var shares []*models.ScriptShare
	if err := cursor.All(ctx, &shares); err != nil {
		return nil, err
	}
	return shares, nil
}

func (r *ScriptShareRepository) Delete(ctx context.Context, scriptID, userID primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"script_id": scriptID,
		"user_id":   userID,
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
