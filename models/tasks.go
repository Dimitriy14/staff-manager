package models

import (
	"time"

	"github.com/google/uuid"
)

type statuses string

const (
	Ready      statuses = "Ready"
	InProgress statuses = "InProgress"
	Done       statuses = "Done"
	Blocked    statuses = "Blocked"
)

type Task struct {
	ID          uuid.UUID `json:"id"`
	Number      uint64    `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   *User     `json:"createdBy,omitempty"`
	UpdatedBy   *User     `json:"updatedBy,omitempty"`
	Assigned    *User     `json:"assigned,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	Status      statuses  `json:"status"`
}

type TaskElastic struct {
	ID          uuid.UUID `json:"id"`
	Number      uint64    `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AssignedID  string    `json:"assignedID"`
	CreatedByID string    `json:"createdByID"`
	UpdatedByID string    `json:"updatedByID"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	Status      statuses  `json:"status"`
}

func (t TaskElastic) IsAssigned() bool {
	return t.AssignedID != ""
}

type TaskSearch struct {
	Search string `json:"search"`
}
