package elasticsearch

import (
	"context"
	"net/http"
	"time"

	"github.com/Dimitriy14/staff-manager/logger"
	"github.com/Dimitriy14/staff-manager/models"

	"github.com/olivere/elastic"
	"github.com/pkg/errors"
)

const elasticSearch = "ElasticSearch"

type Client struct {
	ESClient   *elastic.Client
	BulkProc   *elastic.BulkProcessor
	httpClient *http.Client
	urls       []string
}

type Config struct {
	URLs                 []string   `json:"URLs"`
	MaxIdleConns         int        `json:"MaxIdleConns"`
	IdleConnTimeoutInSec int        `json:"IdleConnTimeoutInSec"`
	ClientTimeoutInSec   int        `json:"ClientTimeoutInSec"`
	BulkConfig           BulkConfig `json:"BulkConfig"`
}

type BulkConfig struct {
	// Name is an optional name to identify this bulk processor.
	Name string `json:"Name"`
	// Workers is the number of concurrent workers allowed to be
	// executed. Defaults to 1 and must be greater or equal to 1.
	Workers int `json:"Workers"`
	// FlushInterval specifies when to flush at the end of the given interval.
	// This is disabled by default. If you want the bulk processor to
	// operate completely asynchronously, set both BulkActions and BulkSize to
	// -1 and set the FlushInterval to a meaningful interval.
	FlushInterval time.Duration `json:"FlushInterval"`
	// BulkSize specifies when to flush based on the size (in bytes) of the actions
	// currently added. Defaults to 5 MB and can be set to -1 to be disabled.
	MaxBulkSize int `json:"MaxBulkSize"`
	// BulkActions specifies when to flush based on the number of actions
	// currently added. Defaults to 1000 and can be set to -1 to be disabled.
	BulkActions int `json:"BulkActions"`
}

func Load(cfg Config, log logger.Logger) (*Client, error) {
	transport := &http.Transport{
		MaxIdleConns:    cfg.MaxIdleConns,
		IdleConnTimeout: time.Duration(cfg.IdleConnTimeoutInSec) * time.Second,
	}
	// setup a http client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.IdleConnTimeoutInSec) * time.Second,
	}

	esClient, err := elastic.NewSimpleClient(
		elastic.SetURL(cfg.URLs...),
		elastic.SetHttpClient(httpClient),
		elastic.SetErrorLog(logger.NewElasticLogger(log)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can't establish connection to ElasticSearch")
	}

	c := &Client{ESClient: esClient, httpClient: httpClient, urls: cfg.URLs}

	c.BulkProc, err = c.startUserBulkProcessor(cfg.BulkConfig, log)
	if err != nil {
		return nil, errors.Wrap(err, "can't start user bulk processor")
	}

	return c, nil
}

func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	if c.ESClient.IsRunning() {
		c.ESClient.Stop()
	}

	return c.BulkProc.Stop()
}

func (c *Client) Status() models.ConnectionStatus {
	var (
		activeNodes = make([]string, 0, len(c.urls))
		downNodes   = make([]string, 0, 0)
	)

	for _, u := range c.urls {
		_, code, err := c.ESClient.Ping(u).Do(context.Background())
		if err != nil || code != http.StatusOK {
			downNodes = append(downNodes, u)
			continue
		}
		activeNodes = append(activeNodes, u)
	}

	return models.ConnectionStatus{
		ServiceName: elasticSearch,
		ActiveNodes: activeNodes,
		DownNodes:   downNodes,
	}
}

func (c *Client) startUserBulkProcessor(cfg BulkConfig, log logger.Logger) (*elastic.BulkProcessor, error) {
	return c.ESClient.BulkProcessor().
		Name(cfg.Name).
		After(AfterCallback(log)).
		Workers(cfg.Workers).
		BulkSize(cfg.MaxBulkSize).
		BulkActions(cfg.BulkActions).
		FlushInterval(cfg.FlushInterval).
		Do(context.Background())
}
