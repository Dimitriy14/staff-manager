package vacation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Dimitriy14/staff-manager/json-validator/schemas"
	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/usecases/vacation"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const layout = "2006-01-02"

func NewService(r *rest.Service, vac vacation.VacationsUsecase, log logger.Logger) *serviceImpl {
	return &serviceImpl{
		r:   r,
		vac: vac,
		log: log,
	}
}

type Service interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetForUser(w http.ResponseWriter, r *http.Request)
	GetMyVacation(w http.ResponseWriter, r *http.Request)
	GetPending(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	CreateNew(w http.ResponseWriter, r *http.Request)
	UpdateStatus(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)
	UpdateExpired(w http.ResponseWriter, r *http.Request)
}

type serviceImpl struct {
	r   *rest.Service
	vac vacation.VacationsUsecase
	log logger.Logger
}

func (s *serviceImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	vacations, err := s.vac.GetAll(ctx)
	if err != nil {
		s.log.Warnf(txID, "GetAll(ctx) err=%s", err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vacations)
}

func (s *serviceImpl) GetForUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		s.log.Warnf(txID, "invalid user id: err=%s", err)
		s.r.SendBadRequest(ctx, w, "invalid user id: err=%s", err)
		return
	}

	vacations, err := s.vac.GetForUser(ctx, uid.String())
	if err != nil {
		s.log.Warnf(txID, "GetByID(ctx, id=%s) err=%s", uid, err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vacations)
}

func (s *serviceImpl) GetMyVacation(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	vacations, err := s.vac.GetForUser(ctx, ua.UserID)
	if err != nil {
		s.log.Warnf(txID, "GetForUser(ctx, id=%s) err=%s", ua.UserID, err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vacations)
}

func (s *serviceImpl) GetPending(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	vacations, err := s.vac.GetPending(ctx)
	if err != nil {
		s.log.Warnf(txID, "GetPending(ctx) err=%s", err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vacations)
}

func (s *serviceImpl) CreateNew(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	body, err := util.RetrieveAndValidate(schemas.VacationCreate, s.log, r)
	if err != nil {
		s.log.Warnf(txID, "validation failed: err=%s", err)
		s.r.SendBadRequest(ctx, w, "validation failed: err=%s", err)
		return
	}

	var vacationReq models.VacationReq
	err = json.Unmarshal(body, &vacationReq)
	if err != nil {
		s.log.Warnf(txID, "unmarshaling failed: err=%s", err)
		s.r.SendBadRequest(ctx, w, "unmarshaling failed: err=%s", err)
		return
	}
	var v models.VacationDB
	v.StartDate, err = time.Parse(layout, vacationReq.StartDate)
	if err != nil {
		s.log.Warnf(txID, "start date parsing failed: err=%s", err)
		s.r.SendBadRequest(ctx, w, "start date parsing failed: err=%s", err)
		return
	}
	v.EndDate, err = time.Parse(layout, vacationReq.EndDate)
	if err != nil {
		s.log.Warnf(txID, "EndDate parsing failed: err=%s", err)
		s.r.SendBadRequest(ctx, w, "EndDate parsing failed: err=%s", err)
		return
	}
	v.UserID = ua.UserID

	if v.StartDate.After(v.EndDate) {
		s.log.Warnf(txID, "start date %v cannot be after end date %v", v.StartDate, v.EndDate)
		s.r.SendBadRequest(ctx, w, "start date %v cannot be after end date %v", v.StartDate, v.EndDate)
		return
	}

	vac, err := s.vac.Save(ctx, v)
	if err != nil {
		s.log.Warnf(txID, "vacation.Save(ctx, vacation=%#v) err=%s", v, err)
		if models.IsErrInvalidData(err) {
			s.r.SendBadRequest(ctx, w, "cannot create vacation: %s", err)
			return
		}
		s.r.SendInternalServerError(ctx, w, "vacation saving failed due to: %s", err)
		return
	}

	s.r.RenderJSON(ctx, w, vac)
}

func (s *serviceImpl) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		s.log.Warnf(txID, "invalid user id: err=%s", err)
		s.r.SendBadRequest(ctx, w, "invalid user id: err=%s", err)
		return
	}

	body, err := util.RetrieveAndValidate(schemas.VacationStatusUpdate, s.log, r)
	if err != nil {
		s.log.Warnf(txID, "validation failed: err=%s", err)
		s.r.SendBadRequest(ctx, w, "validation failed: err=%s", err)
		return
	}

	var v models.VacationStatusUpdate
	err = json.Unmarshal(body, &v)
	if err != nil {
		s.log.Warnf(txID, "unmarshaling failed: err=%s", err)
		s.r.SendBadRequest(ctx, w, "unmarshaling failed: err=%s", err)
		return
	}

	vac, err := s.vac.UpdateVacationStatus(ctx, uid, v.Status)
	if err != nil {
		s.log.Warnf(txID, "UpdateVacationStatus(ctx, id=%s, status=%s) err=%s", uid.String(), v.Status, err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vac)
}

func (s *serviceImpl) Cancel(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		s.log.Warnf(txID, "invalid user id: err=%s", err)
		s.r.SendBadRequest(ctx, w, "invalid user id: err=%s", err)
		return
	}

	vac, err := s.vac.UpdateVacationStatus(ctx, uid, models.Canceled)
	if err != nil {
		s.log.Warnf(txID, "UpdateVacationStatus(ctx, id=%s, status=%s) err=%s", uid.String(), models.Canceled, err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vac)
}

func (s *serviceImpl) GetByID(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		s.log.Warnf(txID, "invalid vacation id: err=%s", err)
		s.r.SendBadRequest(ctx, w, "invalid vacation id: err=%s", err)
		return
	}

	vac, err := s.vac.GetByID(ctx, uid)
	if err != nil {
		s.log.Warnf(txID, "GetVacationByID(ctx, id=%s, status=%s) err=%s", uid.String(), models.Canceled, err)
		s.r.SendInternalServerError(ctx, w, "vacation retrieving failed")
		return
	}

	s.r.RenderJSON(ctx, w, vac)
}

func (s *serviceImpl) UpdateExpired(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	go s.vac.SetExpired(ctx)

	s.r.RenderJSON(ctx, w, rest.Message{Message: fmt.Sprintf("vacation status update started with txID = %s", txID)})
}
