package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
)

func LogMiddleware(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				start      = time.Now()
				ctx        = r.Context()
				txID       = transactionID.FromContext(ctx)
				remoteAddr = r.RemoteAddr
			)

			log.Infof(txID, "REST Started: method=%s remote=%s path=%s",
				r.Method, remoteAddr, r.RequestURI)

			next.ServeHTTP(w, r)
			latency := time.Since(start)

			log.Infof(txID, "REST Completed: latency=%s method=%v remote=%v path=%v r=%v took=%v",
				latency, r.Method, remoteAddr, r.URL.Path, r, latency)
		})
	}
}

// AddIDRequestMiddleware sets TransactionID to header and context
func AddIDRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqID = r.Header.Get(string(transactionID.Key))
		if reqID == "" {
			reqID = uuid.New().String()
			r.Header.Add("TransactionID", reqID)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), transactionID.Key, reqID)))
	})
}
