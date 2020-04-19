package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/Dimitriy14/staff-manager/db"
	"github.com/Dimitriy14/staff-manager/elasticsearch"
	"github.com/Dimitriy14/staff-manager/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
)

type Configuration struct {
	ListenURL     string               `json:"ListenURL"`
	URLPrefix     string               `json:"URLPrefix"`
	AWSSecretName string               `json:"AWSSecretName"`
	AWSRegion     string               `json:"AWSRegion"`
	Logger        logger.Config        `json:"Logger"`
	ElasticSearch elasticsearch.Config `json:"ElasticSearch"`
	DB            db.Config
}

func Load(configFile string) (Configuration, error) {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "loading config file")
	}

	var cfg Configuration
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "unmarshalling config")
	}

	cfg.DB, err = GetDBConfig(cfg.AWSRegion, cfg.AWSSecretName)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "getting DB config")
	}
	return cfg, nil
}

func GetDBConfig(awsRegion, secret string) (cfg db.Config, err error) {
	sess, err := session.NewSession()
	if err != nil {
		return db.Config{}, errors.Wrap(err, "creating AWS session")
	}

	svc := secretsmanager.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

	secretValue, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secret),
		VersionId:    nil,
		VersionStage: nil,
	})
	if err != nil {
		return db.Config{}, errors.Wrap(err, "retrieving secret from AWS Secret Manager")
	}

	err = json.Unmarshal([]byte(*secretValue.SecretString), &cfg)
	return
}
