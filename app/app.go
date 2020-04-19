package app

import (
	"log"
	"os"

	"github.com/Dimitriy14/staff-manager/config"
	"github.com/Dimitriy14/staff-manager/db"
	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/logger"
	"github.com/Dimitriy14/staff-manager/web"
	"github.com/Dimitriy14/staff-manager/web/services/health"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

type Components struct {
	Configuration config.Configuration
	Log           logger.Logger
	Postgres      *db.Client
	ElasticSearch *elasticsearch.Client
	shutdowns     []func() error
}

func LoadApplication(cfgFile string, signal chan os.Signal) (c Components, err error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return Components{}, err
	}
	c.Configuration = cfg

	l, err := logger.Load(cfg.Logger)
	if err != nil {
		return Components{}, err
	}
	c.Log = l
	c.shutdowns = append(c.shutdowns, l.Close)

	pg, err := db.Load(cfg.DB, l)
	if err != nil {
		return Components{}, err
	}
	c.Postgres = pg
	c.shutdowns = append(c.shutdowns, pg.Session.Close)

	es, err := elasticsearch.Load(cfg.ElasticSearch, l)
	if err != nil {
		return Components{}, err
	}
	c.ElasticSearch = es
	c.shutdowns = append(c.shutdowns, es.Close)

	restService := rest.NewRestService(l)
	router := web.NewRouter(c.Configuration.URLPrefix,
		web.Services{
			Health: health.GetHealth(cfg.ListenURL, restService, pg, es),
			Rest:   restService,
		})
	server := web.NewServer(cfg.ListenURL, router, signal)
	server.Start()
	c.shutdowns = append(c.shutdowns, server.Stop)

	log.Println("Start listening on: ", c.Configuration.ListenURL)
	l.Debugf("", "Start application with configuration: %+v", c.Configuration)
	return
}

func (c *Components) Stop() {
	if c == nil {
		return
	}

	for _, f := range c.shutdowns {
		if err := f(); err != nil {
			log.Printf("one of shutdowns is failed: %s", err.Error())
		}
	}
}
