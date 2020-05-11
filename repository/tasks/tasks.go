package tasks

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/models"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

const (
	taskIndex = "tasks"
	taskType  = "task"

	assignedAttribute = "assignedID"
)

func NewRepository(es *elasticsearch.Client) *tasksRepo {
	return &tasksRepo{es: es}
}

type tasksRepo struct {
	es *elasticsearch.Client
}

func (r *tasksRepo) GetUserTasks(ctx context.Context, userID string) ([]models.TaskElastic, error) {
	q := elastic.NewMatchQuery(assignedAttribute, userID)
	resp, err := r.es.ESClient.Search(taskIndex).
		Query(q).
		Do(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "searching tasks for user(id=%s)", userID)
	}

	tasks := make([]models.TaskElastic, 0, resp.TotalHits())
	for _, u := range resp.Each(reflect.TypeOf(models.TaskElastic{})) {
		if task, ok := u.(models.TaskElastic); ok {
			tasks = append(tasks, task)
		}
	}

	return tasks, err
}

func (r *tasksRepo) SaveTask(ctx context.Context, task models.TaskElastic) error {
	_, err := r.es.ESClient.Index().
		Index(taskIndex).
		Id(task.ID.String()).
		BodyJson(task).
		Do(ctx)

	return err
}

func (r *tasksRepo) GetTasks(ctx context.Context, amount int) ([]models.TaskElastic, error) {
	q := elastic.NewMatchAllQuery()
	resp, err := r.es.ESClient.Search(taskIndex).
		From(0).
		Size(amount).
		Query(q).
		Do(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "retrieving all tasks")
	}

	tasks := make([]models.TaskElastic, 0, resp.TotalHits())
	for _, u := range resp.Each(reflect.TypeOf(models.TaskElastic{})) {
		if task, ok := u.(models.TaskElastic); ok {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *tasksRepo) GetTaskByID(ctx context.Context, id string) (models.TaskElastic, error) {
	resp, err := r.es.ESClient.Get().
		Index(taskIndex).
		Id(id).
		Do(ctx)
	if err != nil {
		return models.TaskElastic{}, err
	}

	var t models.TaskElastic
	err = json.Unmarshal(resp.Source, &t)
	return t, err
}

func (r *tasksRepo) CountTasks(ctx context.Context) (int64, error) {
	return r.es.ESClient.Count(taskIndex).Do(ctx)
}

func (r *tasksRepo) Search(ctx context.Context, search string) ([]models.TaskElastic, error) {
	return nil, nil
}

func (r *tasksRepo) UpdateTask(ctx context.Context, task models.TaskElastic) error {
	_, err := r.es.ESClient.Update().
		Index(taskIndex).
		Doc(task).
		Id(task.ID.String()).
		Do(ctx)
	return err
}

func (r *tasksRepo) DeleteTask(ctx context.Context, id string) error {
	_, err := r.es.ESClient.Delete().
		Index(taskIndex).
		Id(id).
		Do(ctx)
	return err
}
