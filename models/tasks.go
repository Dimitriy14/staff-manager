package models

import (
	"time"

	"github.com/google/uuid"
)

type Statuses string

const (
	Ready      Statuses = "Ready"
	InProgress Statuses = "InProgress"
	Done       Statuses = "Done"
	Blocked    Statuses = "Blocked"
)

type Task struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   User      `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedBy   User      `json:"updatedBy"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Assigned    User      `json:"assigned"`
	Status      string    `json:"status"`
}

//
// tasks  -get my tasks; post - create new one; put
// tasks/all - all tasks
// tasks/{id} - get
// search task by number and title
