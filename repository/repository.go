package repository

import "github.com/Dimitriy14/staff-manager/models"

type UserRepository interface {
	Save(u models.User) error
}
