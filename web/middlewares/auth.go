package middlewares

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/usecases/auth"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

func AuthMiddleware(log logger.Logger, auth auth.Authentication, service *rest.Service) mux.MiddlewareFunc {
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

			ua, err := getUserUserAccess(ctx, auth, aout, w)
			if err != nil {
				log.Warnf(txID, "cannot find user due to err: %s", err)
				service.SendUnauthorized(ctx, w, "cannot find user due to err: %s", err)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, models.AccessKey, ua)))
		})
	}
}

func getUserUserAccess(ctx context.Context, auth auth.Authentication, aout models.AuthOutput, w http.ResponseWriter) (*models.UserAccess, error) {
	ua, isTokenExpiried, err := auth.GetUserAccess(ctx, aout.AccessToken)
	if err != nil {
		return nil, err
	}

	if !isTokenExpiried {
		return ua, nil
	}

	newTokens, err := auth.RefreshToken(ctx, aout.RefreshToken)
	if err != nil {
		return nil, err
	}

	util.SetSecureTokens(*newTokens, w)

	return getUserUserAccess(ctx, auth, *newTokens, w)
}

func AdminRestriction(log logger.Logger, service *rest.Service) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx  = r.Context()
				txID = transactionID.FromContext(ctx)
				ua   = util.GetUserAccessFromCtx(ctx)
			)

			if !ua.Role.IsAdmin() {
				log.Warnf(txID, "admin access is restricted for user: %s", ua.UserID)
				service.SendUnauthorized(ctx, w, "admin access is restricted for user: %s", ua.UserID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
