package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScriptType string

const (
	ScriptTypePython ScriptType = "python"
	ScriptTypeGolang ScriptType = "golang"
)

type Script struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Content     string             `bson:"content" json:"content"`
	Type        ScriptType         `bson:"type" json:"type"`
	OwnerID     primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type ScriptShare struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ScriptID  primitive.ObjectID `bson:"script_id" json:"script_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type CreateScriptRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description"`
	Content     string     `json:"content" validate:"required"`
	Type        ScriptType `json:"type" validate:"required,oneof=python golang"`
}

type UpdateScriptRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	Type        ScriptType `json:"type" validate:"omitempty,oneof=python golang"`
}

type ShareScriptRequest struct {
	UserID string `json:"user_id" validate:"required"`
}
