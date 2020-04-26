package app

import (
	"log"
	"os"

	awservices "github.com/Dimitriy14/staff-manager/aws"
	"github.com/Dimitriy14/staff-manager/config"
	"github.com/Dimitriy14/staff-manager/db"
	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/logger"
	authUsecase "github.com/Dimitriy14/staff-manager/usecases/auth"
	"github.com/Dimitriy14/staff-manager/web"
	"github.com/Dimitriy14/staff-manager/web/middlewares"
	"github.com/Dimitriy14/staff-manager/web/services/auth"
	"github.com/Dimitriy14/staff-manager/web/services/health"
	"github.com/Dimitriy14/staff-manager/web/services/rest"

	"github.com/aws/aws-sdk-go/aws/session"
)

type Components struct {
	Configuration config.Configuration
	Log           logger.Logger
	Postgres      *db.Client
	ElasticSearch *elasticsearch.Client
	Cognito       *awservices.CognitoProvider
	shutdowns     []func() error
}

func LoadApplication(cfgFile string, signal chan os.Signal) (c Components, err error) {
	sess, err := session.NewSession()
	if err != nil {
		return c, err
	}

	cfg, err := config.Load(cfgFile, sess)
	if err != nil {
		return Components{}, err
	}
	c.Configuration = cfg
	c.Cognito = awservices.GetCognitoProvider(sess, cfg.AWSRegion, cfg.UserPoolID, cfg.ClientID)

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

	authuc := authUsecase.NewAuthUsecase(c.Cognito, l)
	restService := rest.NewRestService(l)
	a := auth.NewAuthService(authUsecase.NewAuthUsecase(c.Cognito, l), restService, l)
	router := web.NewRouter(c.Configuration.URLPrefix,
		web.Services{
			Health:         health.GetHealth(cfg.ListenURL, restService, pg, es),
			Rest:           restService,
			Auth:           a,
			LogMiddleware:  middlewares.LogMiddleware(l),
			TxIDMiddleware: middlewares.AddIDRequestMiddleware,
			AuthMiddleware: middlewares.AuthMiddleware(l, authuc, restService),
		})
	server := web.NewServer(cfg.ListenURL, router, l, signal)
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
