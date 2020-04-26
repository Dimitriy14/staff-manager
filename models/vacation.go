package models

import (
	"time"

	"github.com/google/uuid"
)

type Vacation struct {
	ID         uuid.UUID `json:"id"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	IsApproved bool      `json:"isApproved"`
	Approver   User      `json:"approver"`
}
