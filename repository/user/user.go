package user

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/models"

	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

const (
	elasticIndex = "staff"
	userType     = "user"

	role           = "role"
	userName       = "firstName"
	userSecondName = "lastName"
	position       = "position"
)

func NewRepository(es *elasticsearch.Client) *repo {
	return &repo{es: es}
}

type repo struct {
	es *elasticsearch.Client
}

func (r *repo) GetUserByID(ctx context.Context, id string) (models.User, error) {
	resp, err := r.es.ESClient.Get().
		Index(elasticIndex).
		Id(id).
		Do(ctx)
	if err != nil {
		return models.User{}, err
	}

	var u models.User
	err = json.Unmarshal(resp.Source, &u)
	return u, err
}

func (r *repo) Save(ctx context.Context, u models.User) error {
	_, err := r.es.ESClient.Index().
		Index(elasticIndex).
		BodyJson(u).
		Id(u.ID.String()).
		Do(ctx)

	return err
}

func (r *repo) Update(ctx context.Context, u models.User) error {
	_, err := r.es.ESClient.Update().
		Index(elasticIndex).
		Doc(u).
		Id(u.ID.String()).
		Do(ctx)

	return err
}

func (r *repo) SearchUsers(ctx context.Context, us models.UserSearch) ([]models.User, error) {
	q := elastic.NewBoolQuery()
	strs := strings.Split(strings.TrimRight(us.ByName, " "), " ")
	if us.ByPosition != "" {
		q.Filter(elastic.NewMatchQuery(position, us.ByPosition))
	}

	switch {
	case len(strs) == 1:
		q.Must(elastic.NewMultiMatchQuery(strings.TrimSpace(strs[0]), userName, userSecondName).Type("phrase_prefix"))
	case len(strs) > 1:
		q.Must(
			elastic.NewMultiMatchQuery(strings.TrimSpace(strs[0]), userName, userSecondName).Type("cross_fields"),
			elastic.NewMultiMatchQuery(strings.TrimSpace(strs[1]), userName, userSecondName).Type("phrase_prefix"),
		)
	}

	resp, err := r.es.ESClient.Search().
		Index(elasticIndex).
		Query(q).
		Do(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "searching user by name %s", us.ByName)
	}

	users := make([]models.User, 0, resp.TotalHits())
	for _, u := range resp.Each(reflect.TypeOf(models.User{})) {
		if user, ok := u.(models.User); ok {
			users = append(users, user)
		}
	}
	return users, nil
}

func (r *repo) GetAdmins(ctx context.Context) ([]models.User, error) {
	q := elastic.NewMatchQuery(role, models.AdminRole)
	resp, err := r.es.ESClient.
		Search(elasticIndex).
		Query(q).
		Do(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "searching admins")
	}

	users := make([]models.User, 0, resp.TotalHits())
	for _, u := range resp.Each(reflect.TypeOf(models.User{})) {
		if user, ok := u.(models.User); ok {
			users = append(users, user)
		}
	}
	return users, nil
}
