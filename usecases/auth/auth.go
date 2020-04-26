package auth

import (
	"context"
	"fmt"

	awservices "github.com/Dimitriy14/staff-manager/aws"
	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"

	"github.com/aws/aws-sdk-go/aws"
	cip "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Authentication interface {
	SignUp(ctx context.Context, username string, id uuid.UUID) error
	SignIn(ctx context.Context, username, password string) (aout *models.AuthOutput, err error)
	RequiredPassword(ctx context.Context, username, password, session string) (aout *models.AuthOutput, err error)
	GetUserID(ctx context.Context, token string) (id string, isTokenExpired bool, err error)
	RefreshToken(ctx context.Context, oldToken string) (aout *models.AuthOutput, err error)
}

func NewAuthUsecase(cognito *awservices.CognitoProvider, log logger.Logger) *auth {
	return &auth{
		cognito: cognito,
		log:     log,
	}
}

type auth struct {
	cognito *awservices.CognitoProvider
	log     logger.Logger
}

const (
	passwordAttribute = "PASSWORD"
	usernameAttribute = "USERNAME"
)

func (a *auth) SignUp(ctx context.Context, email string, id uuid.UUID) error {
	txID := transactionID.FromContext(ctx)
	input := &cip.AdminCreateUserInput{
		DesiredDeliveryMediums: []*string{aws.String(cip.DeliveryMediumTypeEmail)},
		UserAttributes: []*cip.AttributeType{
			{
				Name:  aws.String(cip.VerifiedAttributeTypeEmail),
				Value: aws.String(email),
			}, {
				Name:  aws.String(models.IDAttribute),
				Value: aws.String(id.String()),
			},
		},
		UserPoolId: aws.String(a.cognito.UserPoolID),
		Username:   aws.String(email),
	}

	_, err := a.cognito.Provider.AdminCreateUser(input)
	if err != nil {
		a.log.Warnf(txID, "cannot create user in cognito: err=%s", err)
		return err
	}
	return nil
}

func (a *auth) SignIn(ctx context.Context, email, password string) (*models.AuthOutput, error) {
	txID := transactionID.FromContext(ctx)
	input := &cip.InitiateAuthInput{
		AuthFlow: aws.String(cip.AuthFlowTypeUserPasswordAuth),
		AuthParameters: map[string]*string{
			passwordAttribute: aws.String(password),
			usernameAttribute: aws.String(email),
		},
		ClientId: aws.String(a.cognito.ClientID),
	}

	out, err := a.cognito.Provider.InitiateAuth(input)
	if err != nil {
		a.log.Warnf(txID, "cannot initiate auth for user in cognito: err=%s", err)
		if err, ok := err.(*cip.ResourceNotFoundException); ok {
			return nil, models.NewErrNotFound("InitiateAuth: %s", err)
		}
		return nil, err
	}

	if out.ChallengeName != nil && *out.ChallengeName == cip.ChallengeNameTypeNewPasswordRequired {
		return nil, models.NewRequireNewPasswordError(*out.Session)
	}

	if out.AuthenticationResult != nil && out.AuthenticationResult.AccessToken != nil {
		a.log.Infof(txID, "AuthenticationResult: %#v", out.AuthenticationResult)
		return &models.AuthOutput{
			AccessToken:  *out.AuthenticationResult.AccessToken,
			RefreshToken: *out.AuthenticationResult.RefreshToken,
		}, nil
	}
	return nil, errors.New(fmt.Sprintf("Invalid authentication output:%+v", out))
}

func (a *auth) RequiredPassword(ctx context.Context, email, password, session string) (*models.AuthOutput, error) {
	txID := transactionID.FromContext(ctx)
	input := &cip.RespondToAuthChallengeInput{
		ChallengeName: aws.String(cip.ChallengeNameTypeNewPasswordRequired),
		ChallengeResponses: map[string]*string{
			passwordAttribute: aws.String(password),
			usernameAttribute: aws.String(email),
		},
		ClientId: aws.String(a.cognito.ClientID),
		Session:  aws.String(session),
	}

	out, err := a.cognito.Provider.RespondToAuthChallenge(input)
	if err != nil {
		a.log.Warnf(txID, "cannot initiate auth for user in cognito: err=%s", err)
		return nil, err
	}

	if out.AuthenticationResult != nil && out.AuthenticationResult.AccessToken != nil {
		return &models.AuthOutput{
			AccessToken:  *out.AuthenticationResult.AccessToken,
			RefreshToken: *out.AuthenticationResult.RefreshToken,
		}, nil
	}
	return nil, errors.New(fmt.Sprintf("Invalid authentication output:%+v", out))
}

func (a *auth) GetUserID(ctx context.Context, token string) (string, bool, error) {
	txID := transactionID.FromContext(ctx)
	input := &cip.GetUserInput{AccessToken: aws.String(token)}

	out, err := a.cognito.Provider.GetUser(input)
	if err != nil {
		a.log.Warnf(txID, "got an error while retrieving user(%v)", err)
		if _, ok := err.(*cip.NotAuthorizedException); ok {
			return "", true, nil
		}
		return "", false, err
	}

	for _, attribute := range out.UserAttributes {
		if *attribute.Name == models.IDAttribute {
			return *attribute.Value, false, nil
		}
	}

	return "", false, errors.New(fmt.Sprintf("Invalid authentication output:%+v", out))
}

func (a *auth) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthOutput, error) {
	txID := transactionID.FromContext(ctx)
	input := &cip.InitiateAuthInput{
		AuthFlow: aws.String(cip.AuthFlowTypeRefreshTokenAuth),
		AuthParameters: map[string]*string{
			cip.AuthFlowTypeRefreshToken: aws.String(refreshToken),
		},
		ClientId: aws.String(a.cognito.ClientID),
	}

	out, err := a.cognito.Provider.InitiateAuth(input)
	if err != nil {
		a.log.Warnf(txID, "cannot initiate auth for user in cognito: err=%s", err)
		if _, ok := err.(*cip.ResourceNotFoundException); ok {
			return nil, models.NewErrNotFound("InitiateAuth: %s", err)
		}
		return nil, err
	}

	if out.AuthenticationResult != nil && out.AuthenticationResult.AccessToken != nil {
		return &models.AuthOutput{
			AccessToken:  *out.AuthenticationResult.AccessToken,
			RefreshToken: refreshToken,
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("Invalid authentication output:%+v", out))
}
