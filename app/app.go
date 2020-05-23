package app

import (
	"log"
	"os"

	"github.com/Dimitriy14/staff-manager/web/services/vacation"

	awservices "github.com/Dimitriy14/staff-manager/aws"
	"github.com/Dimitriy14/staff-manager/config"
	"github.com/Dimitriy14/staff-manager/db"
	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/logger"
	"github.com/Dimitriy14/staff-manager/repository/recent-action"
	tasksRepo "github.com/Dimitriy14/staff-manager/repository/tasks"
	"github.com/Dimitriy14/staff-manager/repository/user"
	vacationRepo "github.com/Dimitriy14/staff-manager/repository/vacation"
	authUsecase "github.com/Dimitriy14/staff-manager/usecases/auth"
	"github.com/Dimitriy14/staff-manager/usecases/photos"
	tasksuc "github.com/Dimitriy14/staff-manager/usecases/tasks"
	vacationuc "github.com/Dimitriy14/staff-manager/usecases/vacation"
	"github.com/Dimitriy14/staff-manager/web"
	"github.com/Dimitriy14/staff-manager/web/middlewares"
	"github.com/Dimitriy14/staff-manager/web/services/auth"
	"github.com/Dimitriy14/staff-manager/web/services/health"
	recent_changes "github.com/Dimitriy14/staff-manager/web/services/recent-changes"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
	"github.com/Dimitriy14/staff-manager/web/services/tasks"
	userServ "github.com/Dimitriy14/staff-manager/web/services/user"

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

	userRepo := user.NewRepository(c.ElasticSearch)
	authuc := authUsecase.NewAuthUsecase(c.Cognito, l)
	restService := rest.NewRestService(l)
	a := auth.NewAuthService(authUsecase.NewAuthUsecase(c.Cognito, l), restService, userRepo, l)
	photo := photos.NewPhotosUploader(awservices.GetS3Manager(sess, cfg.AWSRegion), cfg.StorageURL, cfg.BucketName)
	uServ := userServ.NewUserService(restService, l, userRepo, authuc, photo)

	vacRepo := vacationRepo.NewVacationRepo(pg)
	recentActionRepo := recent.NewRecentActionRepo(pg)

	vacationUseCase := vacationuc.NewVacationUseCase(vacRepo, userRepo, recentActionRepo, l)

	taskRepository := tasksRepo.NewRepository(es)
	taskuc := tasksuc.NewTaskUsecase(taskRepository, userRepo, recentActionRepo)
	router := web.NewRouter(
		c.Configuration.URLPrefix,
		c.Configuration.OriginHosts,
		web.Services{
			Health:         health.GetHealth(cfg.ListenURL, restService, pg, es),
			Rest:           restService,
			Auth:           a,
			User:           uServ,
			LogMiddleware:  middlewares.LogMiddleware(l),
			TxIDMiddleware: middlewares.AddIDRequestMiddleware,
			AuthMiddleware: middlewares.AuthMiddleware(l, authuc, restService),
			AdminOnly:      middlewares.AdminRestriction(l, restService),
			Task:           tasks.NewTaskService(taskuc, restService, l),
			RecentChanges:  recent_changes.NewService(recentActionRepo, restService, l),
			Vacation:       vacation.NewService(restService, vacationUseCase, l),
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
