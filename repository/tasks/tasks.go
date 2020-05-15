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

	number = "number"
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

func (r *tasksRepo) GetTasks(ctx context.Context, from, size int) ([]models.TaskElastic, error) {
	q := elastic.NewMatchAllQuery()
	s := elastic.NewFieldSort("updatedAt").Desc()
	resp, err := r.es.ESClient.Search(taskIndex).
		From(from).
		Size(size).
		SortBy(s).
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

func (r *tasksRepo) GetNextTaskIndex(ctx context.Context) (int64, error) {
	agg := elastic.NewMaxAggregation().Field(number)
	resp, err := r.es.ESClient.Search().
		Index(taskIndex).
		Aggregation(number, agg).
		Do(ctx)

	if err != nil {
		if elastic.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}

	if resp.TotalHits() == 0 {
		return 0, nil
	}

	max, found := resp.Aggregations.Max(number)
	if !found {
		return 0, errors.New("cannot found number")
	}

	return int64(*max.Value) + 1, nil
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
