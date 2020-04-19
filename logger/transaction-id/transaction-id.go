package transactionID

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// To prevent golint complaint "should not use basic type string as key in context.WithValue"
type key string

const (
	// Key is a identifier for Context
	Key key = "TransactionID"
	// MasterRouterKey is a identifier for Context Starter method
	MasterRouterKey key = "MasterID"
)

const defaultContextID = "###"

// FromContext extracts TransactionID value from HTTP Request' Context
func FromContext(ctx context.Context) (s string) {
	s, _ = ctx.Value(Key).(string)
	if s == "" {
		s = defaultContextID
	}
	if m, _ := ctx.Value(MasterRouterKey).(string); m != "" {
		s = fmt.Sprintf("%s\t%s", s, m)
	}
	return
}

// NewIDContext sets new TransactionID to context
func NewIDContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, Key, uuid.New().String())
}

// AddIDContext sets existed TransactionID to context
func AddIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, Key, id)
}

// AddMasterContext sets MasterID to context
func AddMasterContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, MasterRouterKey, name)
}

// AddIDRequestMiddleware sets TransactionID to header and context
func AddIDRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqID = r.Header.Get(string(Key))
		if reqID == "" {
			reqID = uuid.New().String()
			r.Header.Add("TransactionID", reqID)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), Key, reqID)))
	})
}

// GetTransactionIDFromContext returns transaction UUID from context
func GetTransactionIDFromContext(ctx context.Context) (id string) {
	var ok bool
	id, ok = ctx.Value(Key).(string)
	if !ok || id == "" || id == defaultContextID {
		return ""
	}
	return id
}
