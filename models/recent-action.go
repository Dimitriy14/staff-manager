package models

import (
	"time"

	"github.com/google/uuid"
)

type ChangesType string

const (
	Assignment       ChangesType = "Assignment"
	TaskStatusChange ChangesType = "TaskStatusChange"
	VacationApprove  ChangesType = "VacationApprove"
	VacationRequest  ChangesType = "VacationRequest"
)

// Assignment task, status task, vacation-approve
type RecentChanges struct {
	ID         uuid.UUID   `json:"id"`
	UserID     string      `json:"userID"`
	OwnerID    string      `json:"ownerID"`
	IncidentID uuid.UUID   `json:"incidentID"`
	Type       ChangesType `json:"type"`
	ChangeTime time.Time   `json:"changeTime"`
}
