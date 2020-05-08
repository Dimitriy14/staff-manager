package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	awservices "github.com/Dimitriy14/staff-manager/aws"
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
	OriginHosts   []string             `json:"OriginHosts"`
	Logger        logger.Config        `json:"Logger"`
	ElasticSearch elasticsearch.Config `json:"ElasticSearch"`
	BucketName    string               `json:"BucketName"`
	StorageURL    string
	DB            db.Config
	CognitoConfig
}

type SecretConfig struct {
	Postgres db.Config     `json:"Postgres"`
	Cognito  CognitoConfig `json:"Cognito"`
}

type CognitoConfig struct {
	ClientID   string `json:"ClientID"`
	UserPoolID string `json:"UserPoolID"`
}

func Load(configFile string, sess *session.Session) (Configuration, error) {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "loading config file")
	}

	var cfg Configuration
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return Configuration{}, errors.Wrap(err, "unmarshalling config")
	}

	scfg, err := getSecretConfig(cfg.AWSSecretName, awservices.GetSecretsManager(sess, cfg.AWSRegion))
	if err != nil {
		return Configuration{}, errors.Wrap(err, "getting secret config")
	}

	cfg.DB = scfg.Postgres
	cfg.CognitoConfig = scfg.Cognito
	cfg.StorageURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.BucketName, cfg.AWSRegion)

	return cfg, nil
}

func getSecretConfig(secret string, manager *awservices.SecretsManager) (cfg SecretConfig, err error) {
	secretValue, err := manager.Secret.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secret),
		VersionId:    nil,
		VersionStage: nil,
	})
	if err != nil {
		return SecretConfig{}, errors.Wrap(err, "retrieving secret from AWS Secret Manager")
	}

	err = json.Unmarshal([]byte(*secretValue.SecretString), &cfg)
	return
}
