package awservices

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type CognitoProvider struct {
	Provider   *cognitoidentityprovider.CognitoIdentityProvider
	ClientID   string
	UserPoolID string
}

type SecretsManager struct {
	Secret *secretsmanager.SecretsManager
}

type S3Manager struct {
	*s3manager.Uploader
}

func GetSecretsManager(sess *session.Session, awsRegion string) *SecretsManager {
	return &SecretsManager{Secret: secretsmanager.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})}
}

func GetCognitoProvider(sess *session.Session, awsRegion, userPoolID, clientID string) *CognitoProvider {
	return &CognitoProvider{
		UserPoolID: userPoolID,
		ClientID:   clientID,
		Provider: cognitoidentityprovider.New(sess, &aws.Config{
			Region: aws.String(awsRegion),
		}),
	}
}

func GetS3Manager(sess *session.Session, awsRegion string) *S3Manager {
	newSess := sess.Copy(&aws.Config{
		Region: aws.String(awsRegion),
	})

	return &S3Manager{Uploader: s3manager.NewUploader(newSess)}
}
