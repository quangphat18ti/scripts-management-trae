package repository

import (
	"context"
	"time"

	"scripts-management/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScriptRepository struct {
	collection *mongo.Collection
}

func NewScriptRepository(db *mongo.Database) *ScriptRepository {
	return &ScriptRepository{
		collection: db.Collection("scripts"),
	}
}

func (r *ScriptRepository) Create(ctx context.Context, script *models.Script) error {
	script.CreatedAt = time.Now()
	script.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, script)
	return err
}

func (r *ScriptRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Script, error) {
	var script models.Script
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&script)
	if err != nil {
		return nil, err
	}
	return &script, nil
}

func (r *ScriptRepository) FindByOwnerID(ctx context.Context, ownerID primitive.ObjectID) ([]*models.Script, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var scripts []*models.Script
	if err := cursor.All(ctx, &scripts); err != nil {
		return nil, err
	}
	return scripts, nil
}

func (r *ScriptRepository) Update(ctx context.Context, script *models.Script) error {
	script.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":        script.Name,
			"description": script.Description,
			"content":     script.Content,
			"type":        script.Type,
			"updated_at":  script.UpdatedAt,
		},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": script.ID}, update)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *ScriptRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
