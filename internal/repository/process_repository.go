package repository

import (
	"context"
	"time"

	"scripts-management/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProcessRepository struct {
	collection *mongo.Collection
}

func NewProcessRepository(db *mongo.Database) *ProcessRepository {
	return &ProcessRepository{
		collection: db.Collection("processes"),
	}
}

func (r *ProcessRepository) Create(ctx context.Context, process *models.Process) error {
	_, err := r.collection.InsertOne(ctx, process)
	return err
}

func (r *ProcessRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Process, error) {
	var process models.Process
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&process)
	if err != nil {
		return nil, err
	}
	return &process, nil
}

func (r *ProcessRepository) FindRunningByScriptID(ctx context.Context, scriptID primitive.ObjectID) (*models.Process, error) {
	var process models.Process
	err := r.collection.FindOne(ctx, bson.M{
		"script_id": scriptID,
		"status":    models.ProcessStatusRunning,
	}).Decode(&process)
	if err != nil {
		return nil, err
	}
	return &process, nil
}

// func (r *ProcessRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.ProcessStatus, exitCode *int) error {
// 	update := bson.M{
// 		"$set": bson.M{
// 			"status":   status,
// 			"end_time": time.Now(),
// 		},
// 	}

// 	if exitCode != nil {
// 		update["$set"].(bson.M)["exit_code"] = exitCode
// 	}

// 	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
// 	return err
// }

func (r *ProcessRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Process, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var processes []*models.Process
	if err := cursor.All(ctx, &processes); err != nil {
		return nil, err
	}
	return processes, nil
}

func (r *ProcessRepository) Update(ctx context.Context, process *models.Process) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": process.ID}, process)
	return err
}

func (r *ProcessRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.ProcessStatus, exitCode *int, err string) error {
	update := bson.M{
		"$set": bson.M{
			"status":   status,
			"end_time": primitive.NewDateTimeFromTime(time.Now().UTC()),
		},
	}

	if exitCode != nil {
		update["$set"].(bson.M)["exit_code"] = exitCode
	}

	if err != "" {
		update["$set"].(bson.M)["error"] = err
	}

	_, updateErr := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return updateErr
}
