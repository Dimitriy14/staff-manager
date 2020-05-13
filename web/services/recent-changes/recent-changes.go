package recent_changes

import (
	"net/http"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/repository"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"
)

type Service interface {
	GetRecentChanges(w http.ResponseWriter, r *http.Request)
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
