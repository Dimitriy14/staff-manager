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
	Name        string    `json:"name"`
	SecondName  string    `json:"secondName"`
	Position    string    `json:"position"`
	MobilePhone string    `json:"mobilePhone,omitempty"`
	DateOfBirth string    `json:"dateOfBirth,omitempty"`
}

type UserRegistration struct {
	User
	Credentials
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
