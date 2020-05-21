package db

import (
	"fmt"

	"github.com/Dimitriy14/staff-manager/models"

	"github.com/Dimitriy14/staff-manager/logger"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
)

const postgres = "Postgres"

type Client struct {
	Session *gorm.DB
	addr    string
}

type Config struct {
	Host         string `json:"Host"`
	Port         string `json:"Port"`
	User         string `json:"User"`
	Password     string `json:"Password"`
	DataBaseName string `json:"DataBaseName"`
}

func Load(cfg Config, log logger.Logger) (*Client, error) {
	var (
		postgres = "postgres"
		dbInfo   = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
	)

	url := fmt.Sprintf(dbInfo, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DataBaseName)

	db, err := gorm.Open(postgres, url)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to postgres:")
	}

	db.SetLogger(logger.NewGORMLogger(log))
	db.LogMode(true)

	db.AutoMigrate(&models.RecentChanges{}, &models.VacationDB{})
	return &Client{Session: db, addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)}, nil
}

func (c *Client) Status() models.ConnectionStatus {
	cs := models.ConnectionStatus{
		ServiceName: postgres,
		ActiveNodes: make([]string, 0, 1),
	}

	err := c.Session.DB().Ping()
	if err != nil {
		cs.DownNodes = append(cs.DownNodes, c.addr)
	} else {
		cs.ActiveNodes = append(cs.ActiveNodes, c.addr)
	}

	return cs
}
