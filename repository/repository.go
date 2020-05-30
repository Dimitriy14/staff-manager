package repository

import (
	"context"

	"github.com/Dimitriy14/staff-manager/models"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetAdmins(ctx context.Context) ([]models.User, error)
	Save(ctx context.Context, u models.User) error
	Update(ctx context.Context, u models.User) error
	SearchUsers(ctx context.Context, user models.UserSearch) ([]models.User, error)
}

type RecentActionRepository interface {
	Save(action models.RecentChanges) error
	GetUserChanges(userID string) ([]models.RecentChanges, error)
}

type TaskRepository interface {
	GetUserTasks(ctx context.Context, userID string) ([]models.TaskElastic, error)
	SaveTask(ctx context.Context, task models.TaskElastic) error
	GetTasks(ctx context.Context, from, size int) ([]models.TaskElastic, error)
	GetTaskByID(ctx context.Context, id string) (models.TaskElastic, error)
	GetNextTaskIndex(ctx context.Context) (int64, error)
	Search(ctx context.Context, search string) ([]models.TaskElastic, error)
	SearchForUser(ctx context.Context, search, userID string) ([]models.TaskElastic, error)
	UpdateTask(ctx context.Context, task models.TaskElastic) error
}

type VacationRepository interface {
	Save(ctx context.Context, vacation models.VacationDB) (*models.VacationDB, error)
	Update(ctx context.Context, vacation models.VacationDB) error
	GetAll(ctx context.Context) ([]models.VacationDB, error)
	GetActual(ctx context.Context) ([]models.VacationDB, error)
	GetPending(ctx context.Context) ([]models.VacationDB, error)
	GetForUser(ctx context.Context, userID string) ([]models.VacationDB, error)
	GetByID(ctx context.Context, vacationID string) (*models.VacationDB, error)
	GetPendingForUser(_ context.Context, userID string) ([]models.VacationDB, error)
}
