package recent_changes

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/mux"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/repository"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

type Service interface {
	GetRecentChanges(w http.ResponseWriter, r *http.Request)
	GetRecentChangesForUser(w http.ResponseWriter, r *http.Request)
}

func NewService(repo repository.RecentActionRepository, r *rest.Service, log logger.Logger) *serviceImpl {
	return &serviceImpl{
		repo: repo,
		r:    r,
		log:  log,
	}
}

type serviceImpl struct {
	repo repository.RecentActionRepository
	r    *rest.Service
	log  logger.Logger
}

func (s *serviceImpl) GetRecentChanges(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	rc, err := s.repo.GetUserChanges(ua.UserID)
	if err != nil {
		s.log.Warnf(txID, "GetUserChanges userID=%s failed due to err=%s", ua.UserID, err)
		s.r.SendInternalServerError(ctx, w, "retrieving user changes failed")
		return
	}

	s.r.RenderJSON(ctx, w, rc)
}

func (s *serviceImpl) GetRecentChangesForUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		s.log.Warnf(txID, "cannot parse user ID: %s", err)
		s.r.SendBadRequest(ctx, w, "cannot parse user ID: %s", err)
		return
	}

	rc, err := s.repo.GetUserChanges(uid.String())
	if err != nil {
		s.log.Warnf(txID, "GetRecentChangesForUser userID=%s failed due to err=%s", uid.String(), err)
		s.r.SendInternalServerError(ctx, w, "retrieving user changes failed")
		return
	}

	s.r.RenderJSON(ctx, w, rc)
}
