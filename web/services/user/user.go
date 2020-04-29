package user

import (
	"encoding/json"
	"net/http"

	"github.com/Dimitriy14/staff-manager/util"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/repository"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

type Service interface {
	Search(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
}

func NewUserService(r *rest.Service, log logger.Logger, user repository.UserRepository) *userService {
	return &userService{
		r:    r,
		log:  log,
		user: user,
	}
}

type userService struct {
	r    *rest.Service
	log  logger.Logger
	user repository.UserRepository
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
		id   = util.GetUserIDFromCtx(ctx)
	)

	user, err := u.user.GetUserByID(ctx, id)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", id, err)
		return
	}

	u.r.RenderJSON(ctx, w, user)
}
