package user

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/Dimitriy14/staff-manager/usecases/photos"

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

const (
	maxImageSize int64 = 10 << 24 // max image size is 10MB
	photo              = "photo"
)

var (
	imageExt = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/gif":  true,
		"image/png":  true,
		"image/bmp":  true,
	}
)

type Service interface {
	Search(w http.ResponseWriter, r *http.Request)

	GetCollege(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	GetAdmins(w http.ResponseWriter, r *http.Request)

	AdminUserUpdate(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)

	UploadImage(w http.ResponseWriter, r *http.Request)
}

func NewUserService(r *rest.Service, log logger.Logger, user repository.UserRepository, a auth.Authentication, photo photos.Uploader) *userService {
	return &userService{
		r:     r,
		log:   log,
		user:  user,
		a:     a,
		photo: photo,
	}
}

type userService struct {
	r     *rest.Service
	log   logger.Logger
	user  repository.UserRepository
	a     auth.Authentication
	photo photos.Uploader
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

func (u *userService) GetAdmins(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
	)

	user, err := u.user.GetAdmins(ctx)
	if err != nil {
		u.log.Warnf(txID, "cannot retrieve admins: err=%s", err)
		u.r.SendInternalServerError(ctx, w, "cannot retrieve admins: err=%s", err)
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

	oldUser, err := u.user.GetUserByID(ctx, ua.UserID)
	if err != nil {
		u.log.Warnf(txID, "cannot find user by id(%s): err=%s", ua.UserID, err)
		u.r.SendInternalServerError(ctx, w, "cannot find user by id(%s): err=%s", ua.UserID, err)
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

	oldUser.MobilePhone = user.MobilePhone
	oldUser.DateOfBirth = user.DateOfBirth
	oldUser.Mood = user.Mood
	err = u.user.Update(ctx, oldUser)
	if err != nil {
		u.log.Warnf(txID, "cannot Update user by id(%s): err=%s", ua.UserID, err)
		u.r.SendInternalServerError(ctx, w, "cannot Update user by id(%s): err=%s", ua.UserID, err)
		return
	}

	u.r.RenderJSON(ctx, w, oldUser)
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
	newUser.ImageURL = oldUser.ImageURL

	err = u.a.UpdateUserRole(ctx, newUser.Email, newUser.Role)
	if err != nil {
		u.log.Warnf(txID, "cannot update user role: err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "cannot update user role: err=%s", id, err)
		return
	}

	err = u.user.Update(ctx, newUser)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", id, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", id, err)
		return
	}

	u.r.RenderJSON(ctx, w, newUser)
}

func (u *userService) UploadImage(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		txID = transactionID.FromContext(ctx)
		ua   = util.GetUserAccessFromCtx(ctx)
	)

	err := r.ParseMultipartForm(maxImageSize)
	if err != nil {
		u.log.Warnf(txID, "cannot parse multipart form due to: err=%s", err)
		u.r.SendBadRequest(ctx, w, "cannot parse multipart form due to: err=%s", err)
		return
	}

	file, header, err := r.FormFile(photo)
	if err != nil {
		u.log.Warnf(txID, "cannot retrieve photo from form: err=%s", err)
		u.r.SendBadRequest(ctx, w, "cannot retrieve photo from form: err=%s", err)
		return
	}

	imageContent, err := ioutil.ReadAll(file)
	if err != nil {
		u.log.Warnf(txID, "cannot read photo content: err=%s", err)
		u.r.SendBadRequest(ctx, w, "cannot read photo content: err=%s", err)
		return
	}

	if !imageExt[http.DetectContentType(imageContent)] {
		u.log.Warnf(txID, "invalid file extension: is not an image")
		u.r.SendBadRequest(ctx, w, "invalid file extension: is not an image")
		return
	}

	user, err := u.user.GetUserByID(ctx, ua.UserID)
	if err != nil {
		u.log.Warnf(txID, "cannot search user by id(%s): err=%s", ua.UserID, err)
		u.r.SendInternalServerError(ctx, w, "cannot search user by id(%s): err=%s", ua.UserID, err)
		return
	}

	link, err := u.photo.Upload(ctx, filepath.Ext(header.Filename), imageContent)
	if err != nil {
		u.log.Warnf(txID, "cannot upload user photo: err=%s", err)
		u.r.SendInternalServerError(ctx, w, "cannot upload user photo: err=%s", err)
		return
	}

	user.ImageURL = link
	err = u.user.Update(ctx, user)
	if err != nil {
		u.log.Warnf(txID, "cannot update user id(%s): err=%s", ua.UserID, err)
		u.r.SendInternalServerError(ctx, w, "cannot update user id(%s): err=%s", ua.UserID, err)
		return
	}

	u.r.RenderJSON(ctx, w, user)
}
