package transactionID

import (
	"context"

	"github.com/google/uuid"
)

// To prevent golint complaint "should not use basic type string as key in context.WithValue"
type key string

const (
	// Key is a identifier for Context
	Key key = "TransactionID"
)

const defaultContextID = "###"

// FromContext extracts TransactionID value from HTTP Request' Context
func FromContext(ctx context.Context) string {
	s, ok := ctx.Value(Key).(string)
	if !ok {
		s = uuid.New().String()
	}
	return s
}

// NewIDContext sets new TransactionID to context
func NewIDContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, Key, uuid.New().String())
}

// AddIDContext sets existed TransactionID to context
func AddIDContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, Key, id)
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
