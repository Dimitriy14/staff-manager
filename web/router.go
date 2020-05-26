package web

import (
	"fmt"
	"net/http"

	"github.com/Dimitriy14/staff-manager/web/services/vacation"

	recent_changes "github.com/Dimitriy14/staff-manager/web/services/recent-changes"

	"github.com/Dimitriy14/staff-manager/web/services/tasks"

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
	Task           tasks.Service
	RecentChanges  recent_changes.Service
	Vacation       vacation.Service
	LogMiddleware  mux.MiddlewareFunc
	TxIDMiddleware mux.MiddlewareFunc
	AuthMiddleware mux.MiddlewareFunc
	AdminOnly      mux.MiddlewareFunc
}

func NewRouter(pathPrefix string, originHosts []string, s Services) *mux.Router {
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
	authorisation.Path("/user/photo").HandlerFunc(s.User.UploadImage).Methods(http.MethodPost)
	authorisation.Path("/user/admins").HandlerFunc(s.User.GetAdmins).Methods(http.MethodGet)

	authorisation.Path(fmt.Sprintf("/user/{id:%s}", UUIDPattern)).HandlerFunc(s.User.GetCollege).Methods(http.MethodGet)

	adminOnly := authorisation.Name("admin").Subrouter()
	adminOnly.Use(s.AdminOnly)
	adminOnly.Path(fmt.Sprintf("/user/{id:%s}", UUIDPattern)).HandlerFunc(s.User.AdminUserUpdate).Methods(http.MethodPut)

	authorisation.Path("/task").HandlerFunc(s.Task.GetMyTasks).Methods(http.MethodGet)
	authorisation.Path("/task").HandlerFunc(s.Task.SaveTask).Methods(http.MethodPost)
	authorisation.Path("/task/search").HandlerFunc(s.Task.SearchForUser).Methods(http.MethodPost)
	authorisation.Path("/task/search/all").HandlerFunc(s.Task.Search).Methods(http.MethodPost)
	authorisation.Path(fmt.Sprintf("/task/{id:%s}", UUIDPattern)).HandlerFunc(s.Task.Update).Methods(http.MethodPut)
	authorisation.Path("/task/list").HandlerFunc(s.Task.GetTasks).Methods(http.MethodGet)
	authorisation.Path(fmt.Sprintf("/task/{id:%s}", UUIDPattern)).HandlerFunc(s.Task.GetTaskByID).Methods(http.MethodGet)
	authorisation.Path(fmt.Sprintf("/task/{id:%s}", UUIDPattern)).HandlerFunc(s.Task.DeleteTask).Methods(http.MethodDelete)
	authorisation.Path(fmt.Sprintf("/task/user/{id:%s}", UUIDPattern)).HandlerFunc(s.Task.GetUserTasks).Methods(http.MethodGet)

	authorisation.Path("/recent").HandlerFunc(s.RecentChanges.GetRecentChanges).Methods(http.MethodGet)
	authorisation.Path(fmt.Sprintf("/recent/user/{id:%s}", UUIDPattern)).HandlerFunc(s.RecentChanges.GetRecentChangesForUser).Methods(http.MethodGet)

	authorisation.Path("/vacations").HandlerFunc(s.Vacation.GetMyVacation).Methods(http.MethodGet)
	authorisation.Path("/vacations").HandlerFunc(s.Vacation.CreateNew).Methods(http.MethodPost)
	authorisation.Path(fmt.Sprintf("/vacations/{id:%s}", UUIDPattern)).HandlerFunc(s.Vacation.GetByID).Methods(http.MethodGet)
	authorisation.Path(fmt.Sprintf("/vacations/{id:%s}", UUIDPattern)).HandlerFunc(s.Vacation.Cancel).Methods(http.MethodDelete)
	adminOnly.Path(fmt.Sprintf("/vacations/{id:%s}", UUIDPattern)).HandlerFunc(s.Vacation.UpdateStatus).Methods(http.MethodPut)
	authorisation.Path(fmt.Sprintf("/vacations/user/{id:%s}", UUIDPattern)).HandlerFunc(s.Vacation.GetForUser).Methods(http.MethodGet)
	authorisation.Path("/vacations/pending").HandlerFunc(s.Vacation.GetPending).Methods(http.MethodGet)
	authorisation.Path("/vacations/all").HandlerFunc(s.Vacation.GetAll).Methods(http.MethodGet)

	var corsRouter = mux.NewRouter()
	{
		corsRouter.PathPrefix(pathPrefix).Handler(negroni.New(
			cors.New(cors.Options{
				AllowedOrigins:   originHosts,
				AllowCredentials: true,
				AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
			}),
			negroni.Wrap(router),
		))
	}

	return corsRouter
}
