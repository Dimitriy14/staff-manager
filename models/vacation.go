package models

import (
	"time"

	"github.com/google/uuid"
)

type VacationStatus string

const (
	Pending  = "Pending"
	Approved = "Approved"
	Rejected = "Rejected"
	Canceled = "Canceled"
	Expired  = "Expired"
)

type VacationReq struct {
	StartDate string
	EndDate   string
}

type Vacation struct {
	ID            uuid.UUID      `json:"id"`
	Number        int            `json:"number"`
	User          *User          `json:"user"`
	StartDate     time.Time      `json:"startDate"`
	EndDate       time.Time      `json:"endDate"`
	Status        VacationStatus `json:"status"`
	UpdateTime    time.Time      `json:"updateTime"`
	StatusChanger *User          `json:"statusChanger,omitempty"`
	WasApproved   bool           `json:"wasApproved"`
}

type VacationDB struct {
	ID                    uuid.UUID      `json:"id" gorm:"primary_key"`
	Number                int            `json:"number" gorm:"AUTO_INCREMENT"`
	UserID                string         `json:"userID"`
	UserFullName          string         `json:"userFullName"`
	StartDate             time.Time      `json:"startDate"`
	EndDate               time.Time      `json:"endDate"`
	Status                VacationStatus `json:"status"`
	UpdateTime            time.Time      `json:"updateTime"`
	StatusChangerFullName string         `json:"statusChangerFullName"`
	StatusChangerID       string         `json:"statusChangerID"`
	WasApproved           bool           `json:"wasApproved"`
}

type VacationStatusUpdate struct {
	Status VacationStatus `json:"status"`
}
