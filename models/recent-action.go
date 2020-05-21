package models

import (
	"time"

	"github.com/google/uuid"
)

type ChangesType string

const (
	Assignment           ChangesType = "Assignment"
	TaskStatusChange     ChangesType = "TaskStatusChange"
	TaskDeletion         ChangesType = "TaskDeletion"
	VacationStatusChange ChangesType = "VacationStatusChange"
	VacationRequest      ChangesType = "VacationRequest"
)

// Assignment task, status task, vacation-approve
type RecentChanges struct {
	ID            uuid.UUID   `json:"id"`
	Title         string      `json:"title"`
	IncidentID    uuid.UUID   `json:"incidentID"`
	Type          ChangesType `json:"type"`
	UserName      string      `json:"userName"`
	UserID        string      `json:"userID"`
	OwnerID       string      `json:"ownerID"`
	UpdatedByName string      `json:"updatedByName"`
	UpdatedByID   string      `json:"updatedByID"`
	ChangeTime    time.Time   `json:"changeTime"`
}
