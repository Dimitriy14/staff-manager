package user

import (
	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/olivere/elastic"
)

type repo struct {
	es *elasticsearch.Client
}

func (r *repo) Save(u models.User) error {

	u.ID.ID()
	req := elastic.NewBulkIndexRequest().Index().Type().Id().Doc(&u).
		r.es.BulkProc.Add(req)

	r.es.ESClient.Search().SortBy()
}
