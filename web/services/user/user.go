package user

import (
	"encoding/json"
	"net/http"

	"github.com/Dimitriy14/staff-manager/usecases/auth"

	"github.com/google/uuid"

	"github.com/gorilla/mux"

	"github.com/Dimitriy14/staff-manager/json-validator/schemas"
	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/repository"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

type Service interface {
	Search(w http.ResponseWriter, r *http.Request)

	GetCollege(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)

	AdminUserUpdate(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

func NewUserService(r *rest.Service, log logger.Logger, user repository.UserRepository, a auth.Authentication) *userService {
	return &userService{
		r:    r,
		log:  log,
		user: user,
		a:    a,
	}
}

type userService struct {
	r    *rest.Service
	log  logger.Logger
	user repository.UserRepository
	a    auth.Authentication
}

func (u *userService) Search(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		us   models.UserSearch
	)

	err := json.NewDecoder(r.Body).Decode(&us)
	if err != nil {
		u.log.Warnf(txID, "cannot decode user search: err=%s", err)
		u.r.SendBadRequest(ctx, w, "invalid user search payload: %s", err)
		return
	}

	users, err := u.user.SearchUsers(ctx, us)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by name(%s): err=%s", us.ByName, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by name(%s): err=%s", us.ByName, err)
		return
	}

	u.r.RenderJSON(ctx, w, users)
}

func (u *userService) GetUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	user, err := u.user.GetUserByID(ctx, ua.UserID)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", ua.UserID, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", ua.UserID, err)
		return
	}

	u.r.RenderJSON(ctx, w, user)
}

// GetCollege in perspective should not retrieve all information
func (u *userService) GetCollege(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	user, err := u.user.GetUserByID(ctx, id)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", id, err)
		return
	}

	u.r.RenderJSON(ctx, w, user)

}

func (u *userService) Update(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	if ua.Role.IsAdmin() {
		u.adminUpdate(w, r, ua.UserID)
		return
	}

	body, err := util.RetrieveAndValidate(schemas.UserUpdate, u.log, r)
	if err != nil {
		u.log.Warnf(txID, "invalid user update payload: err=%s", err)
		u.r.SendBadRequest(ctx, w, "invalid user update payload: err=%s", err)
		return
	}

	var user models.UserUpdate
	err = json.Unmarshal(body, &user)
	if err != nil {
		u.log.Warnf(txID, "invalid user update payload, cannot unmarshal: err=%s", err)
		u.r.SendBadRequest(ctx, w, "invalid user update payload, cannot unmarshal: err=%s", err)
		return
	}

	user.ID = ua.UserID
	err = u.user.Update(ctx, user)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", ua.UserID, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", ua.UserID, err)
		return
	}

	u.r.RenderJSON(ctx, w, user)
}

func (u *userService) AdminUserUpdate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	u.adminUpdate(w, r, id)
}

func (u *userService) adminUpdate(w http.ResponseWriter, r *http.Request, id string) {
	var (
		ctx     = r.Context()
		txID    = transactionID.FromContext(ctx)
		newUser models.User
	)

	userID, err := uuid.Parse(id)
	if err != nil {
		u.log.Warnf(txID, "invalid user id(%s): err=%s", id, err)
		u.r.SendBadRequest(ctx, w, "invalid user id(%s): err=%s", id, err)
		return
	}

	oldUser, err := u.user.GetUserByID(ctx, id)
	if err != nil {
		u.log.Warnf(txID, "GetUserByID id(%s): err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "GetUserByID id(%s): err=%s", id, err)
		return
	}

	body, err := util.RetrieveAndValidate(schemas.AdminUserUpdate, u.log, r)
	if err != nil {
		u.log.Warnf(txID, "invalid user update payload: err=%s", err)
		u.r.SendBadRequest(ctx, w, "invalid user update payload: err=%s", err)
		return
	}

	err = json.Unmarshal(body, &newUser)
	if err != nil {
		u.log.Warnf(txID, "invalid user update payload, cannot unmarshal: err=%s", err)
		u.r.SendBadRequest(ctx, w, "invalid user update payload, cannot unmarshal: err=%s", err)
		return
	}
	newUser.ID = userID
	newUser.Email = oldUser.Email

	err = u.a.UpdateUserRole(ctx, newUser.Email, newUser.Role)
	if err != nil {
		u.log.Warnf(txID, "cannot update user role: err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "cannot update user role: err=%s", id, err)
		return
	}

	err = u.user.AdminUpdate(ctx, newUser)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", id, err)
		return
	}

	u.r.RenderJSON(ctx, w, newUser)
}