package web

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"
	"github.com/urfave/negroni"

	"github.com/Dimitriy14/staff-manager/web/services/auth"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
	"github.com/Dimitriy14/staff-manager/web/services/user"

	"github.com/gorilla/mux"
)

const (
	// UUIDPattern a pattern for UUID matchers
	UUIDPattern = `(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`
)

type Services struct {
	Health         http.HandlerFunc
	Rest           *rest.Service
	Auth           auth.Service
	User           user.Service
	LogMiddleware  mux.MiddlewareFunc
	TxIDMiddleware mux.MiddlewareFunc
	AuthMiddleware mux.MiddlewareFunc
	AdminOnly      mux.MiddlewareFunc
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

	authorisation.Path("/user/search").HandlerFunc(s.User.Search).Methods(http.MethodPost)
	authorisation.Path("/user").HandlerFunc(s.User.GetUser).Methods(http.MethodGet)
	authorisation.Path("/user").HandlerFunc(s.User.Update).Methods(http.MethodPut)

	authorisation.Path(fmt.Sprintf("/user/{id:%s}", UUIDPattern)).HandlerFunc(s.User.GetCollege).Methods(http.MethodGet)

	adminOnly := authorisation.Name("admin").Subrouter()
	adminOnly.Use(s.AdminOnly)
	adminOnly.Path(fmt.Sprintf("/user/{id:%s}", UUIDPattern)).HandlerFunc(s.User.AdminUserUpdate).Methods(http.MethodPut)

	var corsRouter = mux.NewRouter()
	{
		corsRouter.PathPrefix(pathPrefix).Handler(negroni.New(
			cors.New(cors.Options{
				AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
			}),
			negroni.Wrap(router),
		))
	}

	return router
}
