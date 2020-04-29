package repository

import (
	"context"

	"github.com/Dimitriy14/staff-manager/models"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (models.UserRegistration, error)
	Save(ctx context.Context, u models.UserRegistration) error
	SearchUsers(ctx context.Context, user models.UserSearch) ([]models.User, error)
}
