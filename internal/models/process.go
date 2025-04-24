package models

import (
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProcessStatus string

const (
	ProcessStatusRunning ProcessStatus = "running"
	ProcessStatusStopped ProcessStatus = "stopped"
	ProcessStatusFailed  ProcessStatus = "failed"
	ProcessStatusSuccess ProcessStatus = "success"
	ProcessStatusError   ProcessStatus = "error"
)

type Process struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ScriptID   primitive.ObjectID `bson:"script_id" json:"script_id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	PID        int                `bson:"pid" json:"pid"`
	Status     ProcessStatus      `bson:"status" json:"status"`
	StartTime  time.Time          `bson:"start_time" json:"start_time"`
	EndTime    *time.Time         `bson:"end_time,omitempty" json:"end_time,omitempty"`
	ExitCode   *int               `bson:"exit_code,omitempty" json:"exit_code,omitempty"`
	OutputPath string             `bson:"output_path,omitempty" json:"output_path,omitempty"`
	Cmd        *exec.Cmd          `bson:"cmd" json:"cmd"`
	Error      string             `bson:"error,omitempty" json:"error,omitempty"`
}

type RunScriptRequest struct {
	Args []string `json:"args,omitempty"`
}
