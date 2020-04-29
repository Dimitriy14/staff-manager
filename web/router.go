package web

import (
	"net/http"

	"github.com/Dimitriy14/staff-manager/web/services/auth"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
	"github.com/Dimitriy14/staff-manager/web/services/user"

	"github.com/gorilla/mux"
)

type Services struct {
	Health         http.HandlerFunc
	Rest           *rest.Service
	Auth           auth.Service
	User           user.Service
	LogMiddleware  mux.MiddlewareFunc
	TxIDMiddleware mux.MiddlewareFunc
	AuthMiddleware mux.MiddlewareFunc
}

func NewRouter(pathPrefix string, s Services) *mux.Router {
	router := mux.NewRouter().StrictSlash(true).PathPrefix(pathPrefix).Subrouter()
	router.Use(s.TxIDMiddleware, s.LogMiddleware)

	authorisation := router.Name("auth").Subrouter()
	authorisation.Use(s.AuthMiddleware)

	router.Path("/health").HandlerFunc(s.Health).Methods(http.MethodGet)

	router.Path("/signup").HandlerFunc(s.Auth.SignUp).Methods(http.MethodPost)
	router.Path("/signin").HandlerFunc(s.Auth.SignIn).Methods(http.MethodPost)
	router.Path("/password/required").HandlerFunc(s.Auth.RequiredPassword).Methods(http.MethodPost)
	authorisation.Path("/signout").HandlerFunc(s.Auth.SignOut).Methods(http.MethodPost)

	authorisation.Path("/user").HandlerFunc(s.User.GetUser).Methods(http.MethodGet)
	authorisation.Path("/user/search").HandlerFunc(s.User.Search).Methods(http.MethodPost)
	return router
}
