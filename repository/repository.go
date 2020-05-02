package repository

import (
	"context"

	"github.com/Dimitriy14/staff-manager/models"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (models.User, error)
	Save(ctx context.Context, u models.User) error
	Update(ctx context.Context, u models.UserUpdate) error
	AdminUpdate(ctx context.Context, u models.User) error
	SearchUsers(ctx context.Context, user models.UserSearch) ([]models.User, error)
}
