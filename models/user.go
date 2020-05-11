package models

import (
	"github.com/google/uuid"
)

type Key string
type Role string

const (
	AccessKey Key = "user-access"

	IDAttribute    = "custom:id"
	RoleAttribute  = "custom:role"
	EmailAttribute = "email"

	AdminRole Role = "admin"
	UserRole  Role = "user"

	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
	CookiePath   = "/staff/"
)

type AuthOutput struct {
	AccessToken  string
	RefreshToken string
}

type User struct {
	ID          uuid.UUID `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Position    string    `json:"position"`
	MobilePhone string    `json:"mobilePhone,omitempty"`
	DateOfBirth string    `json:"dateOfBirth,omitempty"`
	ImageURL    string    `json:"imageURL,omitempty"`
	Role        Role      `json:"role"`
	Mood        string    `json:"mood"`
	Credentials
}

type UserUpdate struct {
	ID          string `json:"id"`
	MobilePhone string `json:"mobilePhone,omitempty"`
	DateOfBirth string `json:"dateOfBirth,omitempty"`
	ImageURL    string `json:"imageURL,omitempty"`
}

type UserAccess struct {
	Email  string `json:"email"`
	UserID string `json:"userID"`
	Role   Role   `json:"role"`
}

type UserSearch struct {
	ByName     string `json:"name"`
	ByPosition string `json:"position"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

func (r Role) IsAdmin() bool {
	return r == AdminRole
}
