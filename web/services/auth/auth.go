package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Dimitriy14/staff-manager/repository"

	"github.com/Dimitriy14/staff-manager/json-validator/schemas"
	"github.com/Dimitriy14/staff-manager/util"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/usecases/auth"
	"github.com/Dimitriy14/staff-manager/web/services/rest"

	"github.com/google/uuid"
)

const (
	sessionCookie     = "session"
	sessionExpiration = time.Minute * 5
)

type Service interface {
	SignUp(w http.ResponseWriter, r *http.Request)
	SignIn(w http.ResponseWriter, r *http.Request)
	RequiredPassword(w http.ResponseWriter, r *http.Request)

	SignOut(w http.ResponseWriter, r *http.Request)
}

func NewAuthService(authentication auth.Authentication, r *rest.Service, user repository.UserRepository, log logger.Logger) *authService {
	return &authService{
		authentication: authentication,
		r:              r,
		log:            log,
		user:           user,
	}
}

type authService struct {
	authentication auth.Authentication
	r              *rest.Service
	log            logger.Logger
	user           repository.UserRepository
}

func (a *authService) SignUp(w http.ResponseWriter, r *http.Request) {
	var (
		u    models.User
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	body, err := util.RetrieveAndValidate(schemas.UserRegistration, a.log, r)
	if err != nil {
		a.log.Warnf(txID, "cannot retrieve user: err=%s", err)
		a.r.SendBadRequest(ctx, w, "invalid user payload: %s", err)
		return
	}

	err = json.Unmarshal(body, &u)
	if err != nil {
		a.log.Warnf(txID, "cannot unmarshal user: err=%s", err)
		a.r.SendBadRequest(ctx, w, "invalid user payload: %s", err)
		return
	}
	u.ID = uuid.New()

	err = a.authentication.SignUp(ctx, u)
	if err != nil {
		a.log.Warnf(txID, "cannot sign up user due to: err=%s", err)
		a.r.SendInternalServerError(ctx, w, "cannot sign up user due to: err=%s", err)
		return
	}

	err = a.user.Save(ctx, u)
	if err != nil {
		a.log.Warnf(txID, "cannot save user due to: err=%s", err)
		a.r.SendInternalServerError(ctx, w, "cannot save user due to: err=%s", err)
		return
	}

	a.r.RenderJSON(ctx, w, u)
}

func (a *authService) SignIn(w http.ResponseWriter, r *http.Request) {
	var (
		cred models.Credentials
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	body, err := util.RetrieveAndValidate(schemas.SignIn, a.log, r)
	if err != nil {
		a.log.Warnf(txID, "cannot retrieve user: err=%s", err)
		a.r.SendBadRequest(ctx, w, "invalid user payload: %s", err)
		return
	}

	err = json.Unmarshal(body, &cred)
	if err != nil {
		a.log.Warnf(txID, "cannot decode credentials: err=%s", err)
		a.r.SendBadRequest(ctx, w, "invalid credentials payload: %s", err)
		return
	}

	aout, err := a.authentication.SignIn(ctx, cred.Email, cred.Password)
	if err != nil {
		a.log.Warnf(txID, "SignIn for user(%s) was failed due to: %s", cred.Email, err)
		switch e := err.(type) {
		case *models.ErrNotFound:
			a.r.SendNotFound(ctx, w, e.Error())
		case *models.RequireNewPasswordError:
			a.passwordChangeRequired(ctx, w, e.Session)
		default:
			a.r.SendInternalServerError(ctx, w, "cannot sign in due to: %s", err)
		}
		return
	}

	a.successfullyAuthorised(ctx, w, aout)
}

func (a *authService) passwordChangeRequired(ctx context.Context, w http.ResponseWriter, sess string) {
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   sess,
		Path:    models.CookiePath,
		Expires: time.Now().Add(sessionExpiration),
	})

	a.r.RenderJSON(ctx, w, rest.Message{Message: "New password is required"})
}

func (a *authService) successfullyAuthorised(ctx context.Context, w http.ResponseWriter, aout *models.AuthOutput) {
	ua, _, err := a.authentication.GetUserAccess(ctx, aout.AccessToken)
	if err != nil {
		a.log.Errorf(transactionID.FromContext(ctx), "Access token was failed")
		a.r.SendInternalServerError(ctx, w, "cannot sign in due to: %s", err)
		return
	}

	util.SetSecureTokens(*aout, w)
	a.r.RenderJSON(ctx, w, ua)
}

func (a *authService) RequiredPassword(w http.ResponseWriter, r *http.Request) {
	var (
		cred models.Credentials
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	body, err := util.RetrieveAndValidate(schemas.SignIn, a.log, r)
	if err != nil {
		a.log.Warnf(txID, "cannot retrieve user: err=%s", err)
		a.r.SendBadRequest(ctx, w, "invalid user payload: %s", err)
		return
	}

	err = json.Unmarshal(body, &cred)
	if err != nil {
		a.log.Warnf(txID, "cannot decode credentials: err=%s", err)
		a.r.SendBadRequest(ctx, w, "invalid credentials payload: %s", err)
		return
	}

	c, err := r.Cookie(sessionCookie)
	if err != nil {
		a.log.Warnf(txID, "session is not specified: err=%s", err)
		a.r.SendBadRequest(ctx, w, "session is not specified: %s", err)
		return
	}

	token, err := a.authentication.RequiredPassword(ctx, cred.Email, cred.Password, c.Value)
	if err != nil {
		a.log.Warnf(txID, "cannot initiate auth for user in cognito: err=%s", err)
		a.r.SendInternalServerError(ctx, w, "cannot initiate auth for user in cognito: %s", err)
		return
	}

	a.successfullyAuthorised(ctx, w, token)
}

func (a *authService) SignOut(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)

	util.SetSecureTokens(models.AuthOutput{}, w)

	a.r.RenderJSON(ctx, w, struct{}{})
}
