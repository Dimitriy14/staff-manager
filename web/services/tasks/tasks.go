package tasks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Dimitriy14/staff-manager/json-validator/schemas"
	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/usecases/tasks"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/Dimitriy14/staff-manager/web/services/rest"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	from = "from"
	size = "size"
)

func NewTaskService(taskuc tasks.TaskUsecase, r *rest.Service, log logger.Logger) *taskService {
	return &taskService{
		taskuc: taskuc,
		r:      r,
		log:    log,
	}
}

type taskService struct {
	taskuc tasks.TaskUsecase
	r      *rest.Service

	log logger.Logger
}

type Service interface {
	GetUserTasks(w http.ResponseWriter, r *http.Request)
	GetMyTasks(w http.ResponseWriter, r *http.Request)
	SaveTask(w http.ResponseWriter, r *http.Request)
	GetTasks(w http.ResponseWriter, r *http.Request)
	GetTaskByID(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)
	SearchForUser(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	DeleteTask(w http.ResponseWriter, r *http.Request)
}

func (ts *taskService) GetUserTasks(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		ts.log.Warnf(txID, "invalid user id: err=%s", err)
		ts.r.SendBadRequest(ctx, w, "invalid user id: err=%s", err)
		return
	}

	t, err := ts.taskuc.GetUserTasks(ctx, uid.String())
	if err != nil {
		ts.log.Warnf(txID, "GetUserTasks userID=%s failed due to err=%s", id, err)
		ts.r.SendInternalServerError(ctx, w, "user tasks retrieving failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) GetMyTasks(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	t, err := ts.taskuc.GetUserTasks(ctx, ua.UserID)
	if err != nil {
		ts.log.Warnf(txID, "GetMyTasks userID=%s failed due to err=%s", ua.UserID, err)
		ts.r.SendInternalServerError(ctx, w, "user tasks retrieving failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) SaveTask(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	body, err := util.RetrieveAndValidate(schemas.TaskCreation, ts.log, r)
	if err != nil {
		ts.log.Warnf(txID, "invalid task create payload: err=%s", err)
		ts.r.SendBadRequest(ctx, w, "invalid task create payload: err=%s", err)
		return
	}

	var task models.TaskElastic
	err = json.Unmarshal(body, &task)
	if err != nil {
		ts.log.Warnf(txID, "cannot unmarshal task: err=%s", err)
		ts.r.SendBadRequest(ctx, w, "cannot unmarshal task: err=%s", err)
		return
	}
	task.ID = uuid.New()
	task.CreatedByID = ua.UserID
	task.UpdatedByID = ua.UserID
	task.CreatedAt = time.Now().UTC()
	task.UpdatedAt = time.Now().UTC()

	t, err := ts.taskuc.SaveTask(ctx, task)
	if err != nil {
		ts.log.Warnf(txID, "SaveTask userID=%s failed due to err=%s", ua.UserID, err)
		if models.IsErrNotFound(err) {
			ts.r.SendNotFound(ctx, w, "assigned user is not found")
			return
		}
		ts.r.SendInternalServerError(ctx, w, "tasks saving failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) GetTasks(w http.ResponseWriter, r *http.Request) {
	var (
		ctx   = r.Context()
		txID  = transactionID.FromContext(ctx)
		ua    = util.GetUserAccessFromCtx(ctx)
		start = r.URL.Query()[from]
		end   = r.URL.Query()[size]
	)

	if len(start) < 1 || len(end) < 1 {
		ts.log.Warnf(txID, "amountTasks is missing")
		ts.r.SendBadRequest(ctx, w, "amountTasks is missing")
		return
	}

	startFrom, err := strconv.ParseUint(start[0], 10, 64)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse \"from\" number: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse \"from\" number: %s", err)
		return
	}

	size, err := strconv.ParseUint(end[0], 10, 64)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse \"size\" number: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse \"size\" number: %s", err)
		return
	}

	t, err := ts.taskuc.GetTasks(ctx, int(startFrom), int(size))
	if err != nil {
		ts.log.Warnf(txID, "GetTasks userID=%s failed due to err=%s", ua.UserID, err)
		ts.r.SendInternalServerError(ctx, w, "tasks retrieving failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse task ID: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse task ID: %s", err)
		return
	}

	task, err := ts.taskuc.GetTaskByID(ctx, uid)
	if err != nil {
		ts.log.Warnf(txID, "GetTaskByID taskID=%s failed due to err=%s", uid.String(), err)
		ts.r.SendInternalServerError(ctx, w, "tasks retrieving failed")
		return
	}

	ts.r.RenderJSON(ctx, w, task)
}

func (ts *taskService) Search(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		s    models.TaskSearch
	)

	body, err := util.RetrieveAndValidate(schemas.TaskSearch, ts.log, r)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse task search: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse task search: %s", err)
		return
	}

	err = json.Unmarshal(body, &s)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse task search: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse task search: %s", err)
		return
	}

	t, err := ts.taskuc.Search(ctx, s.Search)
	if err != nil {
		ts.log.Warnf(txID, "Search failed due to err=%s", err)
		ts.r.SendInternalServerError(ctx, w, "tasks search failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) SearchForUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
		s    models.TaskSearch
	)

	body, err := util.RetrieveAndValidate(schemas.TaskSearch, ts.log, r)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse task search: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse task search: %s", err)
		return
	}

	err = json.Unmarshal(body, &s)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse task search: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse task search: %s", err)
		return
	}

	t, err := ts.taskuc.SearchForUser(ctx, s.Search, ua.UserID)
	if err != nil {
		ts.log.Warnf(txID, "Search failed due to err=%s", err)
		ts.r.SendInternalServerError(ctx, w, "tasks search failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) Update(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		ts.log.Warnf(txID, "invalid task id: err=%s", err)
		ts.r.SendBadRequest(ctx, w, "invalid task id: err=%s", err)
		return
	}

	body, err := util.RetrieveAndValidate(schemas.TaskUpdate, ts.log, r)
	if err != nil {
		ts.log.Warnf(txID, "invalid task create payload: err=%s", err)
		ts.r.SendBadRequest(ctx, w, "invalid task create payload: err=%s", err)
		return
	}

	var task models.TaskElastic
	err = json.Unmarshal(body, &task)
	if err != nil {
		ts.log.Warnf(txID, "cannot unmarshal task: err=%s", err)
		ts.r.SendBadRequest(ctx, w, "cannot unmarshal task: err=%s", err)
		return
	}
	task.ID = uid
	task.UpdatedByID = ua.UserID
	task.UpdatedAt = time.Now().UTC()

	t, err := ts.taskuc.Update(ctx, task)
	if err != nil {
		ts.log.Warnf(txID, "SaveTask userID=%s failed due to err=%s", ua.UserID, err)
		ts.r.SendInternalServerError(ctx, w, "tasks updating failed")
		return
	}

	ts.r.RenderJSON(ctx, w, t)
}

func (ts *taskService) DeleteTask(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
		id   = mux.Vars(r)["id"]
	)

	uid, err := uuid.Parse(id)
	if err != nil {
		ts.log.Warnf(txID, "cannot parse task ID: %s", err)
		ts.r.SendBadRequest(ctx, w, "cannot parse task ID: %s", err)
		return
	}

	err = ts.taskuc.DeleteTask(ctx, uid, ua.UserID)
	if err != nil {
		ts.log.Warnf(txID, "DeleteTask taskID=%s failed due to err=%s", uid.String(), err)
		if models.IsErrNotFound(err) {
			ts.r.SendNotFound(ctx, w, "task with id=%s is not found")
			return
		}
		ts.r.SendInternalServerError(ctx, w, "tasks saving failed")
		return
	}

	ts.r.SendNoContent(w)
}
