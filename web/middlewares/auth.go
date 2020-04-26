package middlewares

import (
	"context"
	"net/http"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/usecases/auth"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

func AuthMiddleware(log logger.Logger, auth auth.Authentication, service *rest.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx  = r.Context()
				txID = transactionID.FromContext(ctx)
			)

			aout, err := util.RetrieveSecureTokens(r)
			if err != nil {
				log.Warnf(txID, "secure tokens are missing: %s", err)
				service.SendUnauthorized(ctx, w, "secure tokens are missing: %s", err)
				return
			}

			id, err := getUserID(ctx, auth, aout, w)
			if err != nil {
				log.Warnf(txID, "cannot find user due to err: %s", err)
				service.SendUnauthorized(ctx, w, "cannot find user due to err: %s", err)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, models.IDAttribute, id)))
		})
	}
}

func getUserID(ctx context.Context, auth auth.Authentication, aout models.AuthOutput, w http.ResponseWriter) (string, error) {
	id, isTokenExpiried, err := auth.GetUserID(ctx, aout.AccessToken)
	if err != nil {
		return "", err
	}

	if !isTokenExpiried {
		return id, nil
	}

	newTokens, err := auth.RefreshToken(ctx, aout.RefreshToken)
	if err != nil {
		return "", err
	}

	util.SetSecureTokens(*newTokens, w)

	return getUserID(ctx, auth, *newTokens, w)
}
