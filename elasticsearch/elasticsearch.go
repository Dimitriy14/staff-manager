package elasticsearch

import (
	"context"
	"net/http"
	"time"

	"github.com/Dimitriy14/staff-manager/logger"
	"github.com/Dimitriy14/staff-manager/models"
	elastic "github.com/olivere/elastic/v7"

	"github.com/pkg/errors"
)

const elasticSearch = "ElasticSearch"

type Client struct {
	ESClient   *elastic.Client
	httpClient *http.Client
	urls       []string
}

type Config struct {
	URLs                 []string `json:"URLs"`
	MaxIdleConns         int      `json:"MaxIdleConns"`
	IdleConnTimeoutInSec int      `json:"IdleConnTimeoutInSec"`
	ClientTimeoutInSec   int      `json:"ClientTimeoutInSec"`
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

	return c, nil
}

func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	if c.ESClient.IsRunning() {
		c.ESClient.Stop()
	}

	return nil
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
