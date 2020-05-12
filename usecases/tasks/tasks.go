package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/repository"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type TaskUsecase interface {
	GetUserTasks(ctx context.Context, userID string) ([]models.Task, error)
	SaveTask(ctx context.Context, task models.TaskElastic) (models.Task, error)
	GetTasks(ctx context.Context, from, size int) ([]models.Task, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (models.Task, error)
	Search(ctx context.Context, search string) ([]models.TaskElastic, error)
	Update(ctx context.Context, task models.TaskElastic) (models.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

func NewTaskUsecase(
	taskRepo repository.TaskRepository,
	userRepo repository.UserRepository,
	recentChangesRepo repository.RecentActionRepository) *taskUsecase {
	return &taskUsecase{
		TaskRepository:    taskRepo,
		userRepo:          userRepo,
		recentChangesRepo: recentChangesRepo,
	}
}

const (
	numOfWorker = 3
)

type taskUsecase struct {
	repository.TaskRepository

	userRepo          repository.UserRepository
	recentChangesRepo repository.RecentActionRepository
}

func (u *taskUsecase) SaveTask(ctx context.Context, task models.TaskElastic) (models.Task, error) {
	creatorUser, err := u.userRepo.GetUserByID(ctx, task.CreatedByID)
	if err != nil {
		return models.Task{}, models.NewErrNotFound("creator user with id=%s is not found", task.CreatedByID)
	}

	count, err := u.CountTasks(ctx)
	if err != nil {
		return models.Task{}, errors.Wrap(err, "cannot count tasks")
	}
	task.Number = fmt.Sprintf("%d", count)
	task.Status = models.Ready

	t := copyToTask(task)

	if task.IsAssigned() {
		t.Assigned, err = u.assignedUser(ctx, task.AssignedID, creatorUser.ID.String(), task.ID)
		if err != nil {
			return models.Task{}, err
		}
	}

	err = u.TaskRepository.SaveTask(ctx, task)
	if err != nil {
		return models.Task{}, errors.Wrap(err, "cannot save task")
	}

	t.CreatedBy = &creatorUser
	t.UpdatedBy = &creatorUser
	return t, nil
}

func (u *taskUsecase) assignedUser(ctx context.Context, assignedID, creatorID string, taskID uuid.UUID) (*models.User, error) {
	assignedUser, err := u.userRepo.GetUserByID(ctx, assignedID)
	if err != nil {
		return nil, models.NewErrNotFound("assigned user with id=%s is not found", assignedID)
	}

	return &assignedUser, u.recentChangesRepo.Save(models.RecentChanges{
		ID:         uuid.New(),
		UserID:     assignedUser.ID.String(),
		OwnerID:    creatorID,
		IncidentID: taskID,
		Type:       models.Assignment,
		ChangeTime: time.Now(),
	})
}

func (u *taskUsecase) GetUserTasks(ctx context.Context, userID string) ([]models.Task, error) {
	tasks, err := u.TaskRepository.GetUserTasks(ctx, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot retrieve tasks for userID=%s", userID)
	}

	return u.joinTasks(ctx, tasks...)
}

func (u *taskUsecase) GetTasks(ctx context.Context, from, size int) ([]models.Task, error) {
	tasks, err := u.TaskRepository.GetTasks(ctx, from, size)
	if err != nil {
		return nil, errors.Wrap(err, "cannot retrieve tasks")
	}

	return u.joinTasks(ctx, tasks...)
}

func (u *taskUsecase) GetTaskByID(ctx context.Context, id uuid.UUID) (models.Task, error) {
	task, err := u.TaskRepository.GetTaskByID(ctx, id.String())
	if err != nil {
		return models.Task{}, errors.Wrapf(err, "cannot retrieve task by id=%s", id)
	}

	return u.joinTaskWithUsers(ctx, task)
}

func (u *taskUsecase) Update(ctx context.Context, task models.TaskElastic) (models.Task, error) {
	oldTask, err := u.TaskRepository.GetTaskByID(ctx, task.ID.String())
	if err != nil {
		return models.Task{}, errors.Wrapf(err, "cannot retrieve task by id=%s", task.ID)
	}
	task.CreatedByID = oldTask.CreatedByID
	task.CreatedAt = oldTask.CreatedAt
	task.Number = oldTask.Number

	if oldTask.AssignedID != task.AssignedID && task.IsAssigned() {
		_, err = u.assignedUser(ctx, task.AssignedID, task.CreatedByID, task.ID)
		if err != nil {
			return models.Task{}, err
		}
	}

	t, err := u.joinTaskWithUsers(ctx, task)
	if err != nil {
		return models.Task{}, err
	}

	return t, u.TaskRepository.UpdateTask(ctx, task)
}

func (u *taskUsecase) joinTasks(ctx context.Context, tasks ...models.TaskElastic) ([]models.Task, error) {
	var (
		taskIn  = make(chan models.TaskElastic)
		taskOut = make(chan models.Task)
		errch   = make(chan error)
	)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < numOfWorker; i++ {
		go func() {
			for task := range taskIn {
				joinedTask, err := u.joinTaskWithUsers(ctx, task)
				if err != nil {
					errch <- err
					return
				}

				select {
				case <-ctx.Done():
					return
				case taskOut <- joinedTask:
				}

			}
		}()
	}

	go func() {
		for _, task := range tasks {
			taskIn <- task
		}
		close(taskIn)
	}()

	joinedTasks := make([]models.Task, 0, len(tasks))
	for i := 0; i < len(tasks); i++ {
		select {
		case task := <-taskOut:
			joinedTasks = append(joinedTasks, task)
		case err := <-errch:
			return nil, err
		}
	}

	return joinedTasks, nil
}

func (u *taskUsecase) joinTaskWithUsers(ctx context.Context, task models.TaskElastic) (models.Task, error) {
	var assignedUser *models.User
	if task.IsAssigned() {
		au, err := u.userRepo.GetUserByID(ctx, task.AssignedID)
		if err != nil {
			return models.Task{}, models.NewErrNotFound("assigned user with id=%s is not found", task.AssignedID)
		}
		assignedUser = &au
	}

	creatorUser, err := u.userRepo.GetUserByID(ctx, task.CreatedByID)
	if err != nil {
		return models.Task{}, models.NewErrNotFound("creator user with id=%s is not found", task.CreatedByID)
	}

	updaterUser, err := u.userRepo.GetUserByID(ctx, task.UpdatedByID)
	if err != nil {
		return models.Task{}, models.NewErrNotFound("updater user with id=%s is not found", task.UpdatedByID)
	}

	t := copyToTask(task)
	t.Assigned = assignedUser
	t.CreatedBy = &creatorUser
	t.UpdatedBy = &updaterUser
	return t, nil
}

func (u *taskUsecase) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return u.TaskRepository.DeleteTask(ctx, id.String())
}

func copyToTask(te models.TaskElastic) models.Task {
	return models.Task{
		ID:          te.ID,
		Number:      te.Number,
		Title:       te.Title,
		Description: te.Description,
		UpdatedAt:   te.UpdatedAt,
		CreatedAt:   te.CreatedAt,
		Status:      te.Status,
	}
}
