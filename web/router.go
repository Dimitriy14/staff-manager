package web

import (
	"net/http"

	"github.com/Dimitriy14/staff-manager/web/services/auth"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
	"github.com/gorilla/mux"
)

type Services struct {
	Health         http.HandlerFunc
	Rest           *rest.Service
	Auth           auth.Service
	LogMiddleware  func(http.Handler) http.Handler
	TxIDMiddleware func(http.Handler) http.Handler
	AuthMiddleware func(http.Handler) http.Handler
}

func NewRouter(pathPrefix string, s Services) *mux.Router {
	router := mux.NewRouter().StrictSlash(true).PathPrefix(pathPrefix).Subrouter()
	router.Use(s.TxIDMiddleware, s.LogMiddleware)

	authorisation := router.Name("auth").Subrouter()
	authorisation.Use(s.AuthMiddleware)

	authorisation.Path("/health").HandlerFunc(s.Health).Methods(http.MethodGet)
	router.Path("/signup").HandlerFunc(s.Auth.SignUp).Methods(http.MethodPost)
	router.Path("/signin").HandlerFunc(s.Auth.SignIn).Methods(http.MethodPost)
	router.Path("/password/required").HandlerFunc(s.Auth.RequiredPassword).Methods(http.MethodPost)
	authorisation.Path("/signout").HandlerFunc(s.Auth.SignOut).Methods(http.MethodPost)
	return router
}
