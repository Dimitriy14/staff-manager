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
	"github.com/pkg/errors"
)

type Authentication interface {
	SignUp(ctx context.Context, user models.User) error
	SignIn(ctx context.Context, username, password string) (aout *models.AuthOutput, err error)
	RequiredPassword(ctx context.Context, username, password, session string) (aout *models.AuthOutput, err error)
	GetUserAccess(ctx context.Context, token string) (ua *models.UserAccess, isTokenExpired bool, err error)
	RefreshToken(ctx context.Context, oldToken string) (aout *models.AuthOutput, err error)
	UpdateUserRole(ctx context.Context, email string, role models.Role) error
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
	passwordAttribute    = "PASSWORD"
	newPasswordAttribute = "NEW_PASSWORD"
	usernameAttribute    = "USERNAME"
)

func (a *auth) SignUp(ctx context.Context, user models.User) error {
	var (
		txID = transactionID.FromContext(ctx)
	)

	input := &cip.AdminCreateUserInput{
		DesiredDeliveryMediums: []*string{aws.String(cip.DeliveryMediumTypeEmail)},
		UserAttributes: []*cip.AttributeType{
			{
				Name:  aws.String(cip.VerifiedAttributeTypeEmail),
				Value: aws.String(user.Email),
			}, {
				Name:  aws.String(models.IDAttribute),
				Value: aws.String(user.ID.String()),
			}, {
				Name:  aws.String(models.RoleAttribute),
				Value: aws.String(string(user.Role)),
			},
		},
		UserPoolId: aws.String(a.cognito.UserPoolID),
		Username:   aws.String(user.Email),
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
			newPasswordAttribute: aws.String(password),
			usernameAttribute:    aws.String(email),
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

func (a *auth) GetUserAccess(ctx context.Context, token string) (*models.UserAccess, bool, error) {
	var (
		txID  = transactionID.FromContext(ctx)
		input = &cip.GetUserInput{AccessToken: aws.String(token)}
		ua    models.UserAccess
	)

	out, err := a.cognito.Provider.GetUser(input)
	if err != nil {
		a.log.Warnf(txID, "got an error while retrieving user(%v)", err)
		if _, ok := err.(*cip.NotAuthorizedException); ok {
			return nil, true, nil
		}
		return nil, false, err
	}

	for _, attribute := range out.UserAttributes {
		switch *attribute.Name {
		case models.IDAttribute:
			ua.UserID = *attribute.Value
		case models.RoleAttribute:
			ua.Role = models.Role(*attribute.Value)
		case models.EmailAttribute:
			ua.Email = *attribute.Value
		}
	}

	return &ua, false, nil
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

func (a *auth) UpdateUserRole(ctx context.Context, email string, role models.Role) error {
	var (
		txID = transactionID.FromContext(ctx)
	)

	input := &cip.AdminUpdateUserAttributesInput{
		UserAttributes: []*cip.AttributeType{
			{
				Name:  aws.String(models.RoleAttribute),
				Value: aws.String(string(role)),
			},
		},
		UserPoolId: aws.String(a.cognito.UserPoolID),
		Username:   aws.String(email),
	}

	_, err := a.cognito.Provider.AdminUpdateUserAttributes(input)
	if err != nil {
		a.log.Warnf(txID, "cannot update user role in cognito: err=%s", err)
		return err
	}
	return nil
}
