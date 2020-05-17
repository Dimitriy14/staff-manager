package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
)

func NewRestService(log logger.Logger) *Service {
	return &Service{log: log}
}

type Service struct {
	log logger.Logger
}

const (
	// UUIDPattern a pattern for UUID matchers
	UUIDPattern = `(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`
)

// Message contains the message to send as a response
type Message struct {
	Message string `json:"message"`
}

// RenderJSON is used for rendering JSON response body with appropriate headers
func (r *Service) RenderJSON(ctx context.Context, w http.ResponseWriter, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		r.SendInternalServerError(ctx, w, err.Error())
		return
	}
	r.render(ctx, w, http.StatusOK, data)
}

func (r *Service) render(ctx context.Context, w http.ResponseWriter, code int, response []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		r.log.Warnf(transactionID.FromContext(ctx), "Write request failed, error:%v", err)
	}
	if code < 300 {
		r.log.Infof(transactionID.FromContext(ctx), "Request success, Code=%d", code)
	} else {
		r.log.Warnf(transactionID.FromContext(ctx), "Request failed, Code=%d", code)
	}
}

// SendBadRequest sends Bad Request Status and logs an error if it exists
func (r *Service) SendBadRequest(ctx context.Context, w http.ResponseWriter, message string, v ...interface{}) {
	r.sendMessage(ctx, w, http.StatusBadRequest, message, v...)
}

// SendUnauthorized sends Unauthorized Status and logs an error if it exists
func (r *Service) SendUnauthorized(ctx context.Context, w http.ResponseWriter, message string, v ...interface{}) {
	r.sendMessage(ctx, w, http.StatusUnauthorized, message, v...)
}

// SendNotFound sends Not Fount Status and logs an error if it exists
func (r *Service) SendNotFound(ctx context.Context, w http.ResponseWriter, message string, v ...interface{}) {
	r.sendMessage(ctx, w, http.StatusNotFound, message, v...)
}

// SendInternalServerError sends Internal Server Error Status and logs an error if it exists
func (r *Service) SendInternalServerError(ctx context.Context, w http.ResponseWriter, message string, v ...interface{}) {
	r.sendMessage(ctx, w, http.StatusInternalServerError, message, v...)
}

// SendMessage writes a defined string as an error message
// with appropriate headers to the HTTP response
func (r *Service) sendMessage(ctx context.Context, w http.ResponseWriter, code int, format string, v ...interface{}) {
	errorMessage := Message{Message: fmt.Sprintf(format, v...)}

	data, err := json.Marshal(errorMessage)
	if err != nil {
		r.log.Error(transactionID.FromContext(ctx), "", "Failed to Marshal data, error:%v", err)
	}
	r.render(ctx, w, code, data)
}

// SendNoContent sends to the client an empty response with the 204 (NOCONTENT) status
func (r *Service) SendNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
