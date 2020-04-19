package rest

import (
	"context"
	"encoding/json"
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
	UUIDPattern              = `(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`
	accessControlAllowOrigin = "Access-Control-Allow-Origin"
)

// Message contains the message to send as a response
type Message struct {
	ErrorID string `json:"id,omitempty"`
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

	// HTTP access control (CORS)
	w.Header().Set(accessControlAllowOrigin, "*")

	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		r.log.Warn(transactionID.FromContext(ctx), "Write request failed, error:%v", err)
	}
	if code < 300 {
		r.log.Info(transactionID.FromContext(ctx), "Request success, Code=%d", code)
	} else {
		r.log.Warn(transactionID.FromContext(ctx), "Request failed, Code=%d", code)
	}
}

// SendBadRequest sends Bad Request Status and logs an error if it exists
func (r *Service) SendBadRequest(ctx context.Context, log logger.Logger, w http.ResponseWriter, message string, errorID ...string) {
	r.sendMessage(ctx, w, http.StatusBadRequest, message, errorID...)
}

// SendNotFound sends Not Fount Status and logs an error if it exists
func (r *Service) SendNotFound(ctx context.Context, log logger.Logger, w http.ResponseWriter, message string) {
	r.sendMessage(ctx, w, http.StatusNotFound, message)
}

// SendInternalServerError sends Internal Server Error Status and logs an error if it exists
func (r *Service) SendInternalServerError(ctx context.Context, w http.ResponseWriter, message string) {
	r.sendMessage(ctx, w, http.StatusInternalServerError, message)
}

// SendMessage writes a defined string as an error message
// with appropriate headers to the HTTP response
func (r *Service) sendMessage(ctx context.Context, w http.ResponseWriter, code int, message string, errorID ...string) {
	w.Header().Set(accessControlAllowOrigin, "*")
	errorMessage := Message{Message: message}
	if len(errorID) == 1 {
		errorMessage.ErrorID = errorID[0]
	}
	data, err := json.Marshal(errorMessage)
	if err != nil {
		r.log.Error(transactionID.FromContext(ctx), "", "Failed to Marshal data, error:%v", err)
	}
	r.render(ctx, w, code, data)
}
