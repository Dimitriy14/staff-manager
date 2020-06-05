package util

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Dimitriy14/staff-manager/models"

	jsonvalidator "github.com/Dimitriy14/staff-manager/json-validator"
	"github.com/Dimitriy14/staff-manager/logger"
)

// CloseReqBody closes req.Body with returned error check
func CloseReqBody(log logger.Logger, req *http.Request) {
	if req == nil || req.Body == nil {
		return
	}
	err := req.Body.Close()
	if err != nil {
		log.Warnf("", "failed to close request body (err: %v)", err)
	}
}

func RetrieveAndValidate(schemaName string, log logger.Logger, req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request should not be equal <nil>")
	}

	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer CloseReqBody(log, req)

	err = jsonvalidator.Validate(schemaName, content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func SetSecureTokens(aout models.AuthOutput, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     models.AccessToken,
		Value:    aout.AccessToken,
		Path:     models.CookiePath,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     models.RefreshToken,
		Value:    aout.RefreshToken,
		Path:     models.CookiePath,
		HttpOnly: true,
	})
}

func RetrieveSecureTokens(req *http.Request) (models.AuthOutput, error) {
	if req == nil {
		return models.AuthOutput{}, errors.New("request should not be equal <nil>")
	}

	at, err := req.Cookie(models.AccessToken)
	if err != nil {
		return models.AuthOutput{}, err
	}

	rt, err := req.Cookie(models.RefreshToken)
	if err != nil {
		return models.AuthOutput{}, err
	}

	return models.AuthOutput{
		AccessToken:  at.Value,
		RefreshToken: rt.Value,
	}, nil
}

func GetUserAccessFromCtx(ctx context.Context) models.UserAccess {
	userID := ctx.Value(models.AccessKey)
	id, ok := userID.(*models.UserAccess)
	if !ok {
		return models.UserAccess{}
	}
	return *id
}
