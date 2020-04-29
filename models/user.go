package models

import (
	"github.com/google/uuid"
)

const (
	IDAttribute = "custom:id"

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
	Credentials
}

type UserSearch struct {
	ByName     string `json:"name"`
	ByPosition string `json:"position"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"-"`
}
