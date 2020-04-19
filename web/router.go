package web

import (
	"net/http"

	"github.com/Dimitriy14/staff-manager/web/services/rest"

	"github.com/gorilla/mux"
)

type Services struct {
	Health http.HandlerFunc
	Rest   *rest.Service
}

func NewRouter(pathPrefix string, s Services) *mux.Router {
	router := mux.NewRouter().StrictSlash(true).PathPrefix(pathPrefix).Subrouter()

	router.Path("/health").HandlerFunc(s.Health).Methods(http.MethodGet)
	return router
}
