package repository

import (
	"context"

	"github.com/Dimitriy14/staff-manager/models"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (models.User, error)
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
	UpdateTask(ctx context.Context, task models.TaskElastic) error
	DeleteTask(ctx context.Context, id string) error
}

//
// tasks  -get my tasks; post - create new one; put
// tasks/all - all tasks
// tasks/{id} - get
// search task by number and title
