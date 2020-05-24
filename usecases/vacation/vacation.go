package vacation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dimitriy14/staff-manager/logger"
	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/repository"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/senseyeio/spaniel"

	"github.com/google/uuid"
)

type VacationsUsecase interface {
	Save(ctx context.Context, vacation models.VacationDB) (*models.Vacation, error)
	UpdateVacationStatus(ctx context.Context, vacationID uuid.UUID, status models.VacationStatus) (*models.Vacation, error)
	GetAll(ctx context.Context) ([]models.Vacation, error)
	GetPending(ctx context.Context) ([]models.Vacation, error)
	GetForUser(ctx context.Context, userID string) ([]models.Vacation, error)
	GetByID(ctx context.Context, vacationID uuid.UUID) (*models.Vacation, error)
}

func NewVacationUseCase(vacationRepo repository.VacationRepository,
	userRepo repository.UserRepository,
	recentChangesRepo repository.RecentActionRepository,
	log logger.Logger) *vacationsUsecase {
	return &vacationsUsecase{
		VacationRepository: vacationRepo,
		userRepo:           userRepo,
		recentChangesRepo:  recentChangesRepo,
		log:                log,
	}
}

type vacationsUsecase struct {
	repository.VacationRepository
	userRepo          repository.UserRepository
	recentChangesRepo repository.RecentActionRepository
	log               logger.Logger
}

func (u *vacationsUsecase) Save(ctx context.Context, vacation models.VacationDB) (*models.Vacation, error) {
	actualVacations, err := u.VacationRepository.GetPendingForUser(ctx, vacation.UserID)
	if err != nil {
		return nil, err
	}

	for _, actualVacation := range actualVacations {
		if isTimeIntersected(actualVacation.StartDate, actualVacation.EndDate, vacation.StartDate, vacation.EndDate) {
			return nil, errors.New(
				fmt.Sprintf(
					"cannot create vacation with start date = %s, reason: intersection with actual vacation with id = %s ends at %s",
					vacation.StartDate, actualVacation.ID, actualVacation.EndDate))
		}
	}

	user, err := u.userRepo.GetUserByID(ctx, vacation.UserID)
	if err != nil {
		return nil, err
	}

	vacation.ID = uuid.New()
	vacation.UserFullName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	vacation.UpdateTime = time.Now().UTC()
	vacation.Status = models.Pending

	createdVacation, err := u.VacationRepository.Save(ctx, vacation)
	if err != nil {
		return nil, err
	}

	vac := copyToVacation(*createdVacation)
	vac.User = &user

	err = u.recentChangesRepo.Save(models.RecentChanges{
		ID:         uuid.New(),
		Title:      fmt.Sprintf("%d Vacation", vac.Number),
		IncidentID: vacation.ID,
		Type:       models.VacationRequest,
		UserName:   vacation.UserFullName,
		UserID:     user.ID.String(),
		OwnerID:    user.ID.String(),
		ChangeTime: vacation.UpdateTime,
		Status:     string(vacation.Status),
	})

	return &vac, err
}

func (u *vacationsUsecase) UpdateVacationStatus(ctx context.Context, vacationID uuid.UUID, status models.VacationStatus) (*models.Vacation, error) {
	userAcces := util.GetUserAccessFromCtx(ctx)
	oldVacation, err := u.VacationRepository.GetByID(ctx, vacationID.String())
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetUserByID(ctx, oldVacation.UserID)
	if err != nil {
		return nil, err
	}

	userAccess := util.GetUserAccessFromCtx(ctx)
	statusChanger, err := u.userRepo.GetUserByID(ctx, userAccess.UserID)
	if err != nil {
		return nil, err
	}

	oldVacation.WasApproved = status == models.Approved
	oldVacation.Status = status
	oldVacation.StatusChangerFullName = fmt.Sprintf("%s %s", statusChanger.FirstName, statusChanger.LastName)
	oldVacation.StatusChangerID = userAcces.UserID
	oldVacation.UpdateTime = time.Now().UTC()
	vac := copyToVacation(*oldVacation)
	vac.User = &user
	vac.StatusChanger = &statusChanger

	_, err = u.VacationRepository.Save(ctx, *oldVacation)
	if err != nil {
		return nil, err
	}

	err = u.recentChangesRepo.Save(models.RecentChanges{
		ID:            uuid.New(),
		Title:         fmt.Sprintf("%d Vacation %s", vac.Number, status),
		IncidentID:    oldVacation.ID,
		Type:          models.VacationStatusChange,
		UserName:      fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		UserID:        user.ID.String(),
		OwnerID:       user.ID.String(),
		UpdatedByName: oldVacation.StatusChangerFullName,
		UpdatedByID:   oldVacation.StatusChangerID,
		ChangeTime:    oldVacation.UpdateTime,
		Status:        string(status),
	})

	return &vac, err
}

func (u *vacationsUsecase) GetByID(ctx context.Context, vacationID uuid.UUID) (*models.Vacation, error) {
	vacDB, err := u.VacationRepository.GetByID(ctx, vacationID.String())
	if err != nil {
		return nil, err
	}
	vacation := copyToVacation(*vacDB)

	user, err := u.userRepo.GetUserByID(ctx, vacDB.UserID)
	if err != nil {
		return nil, err
	}

	if vacDB.StatusChangerID != "" {
		changer, err := u.userRepo.GetUserByID(ctx, vacDB.UserID)
		if err != nil {
			return nil, err
		}
		vacation.StatusChanger = &changer
	}
	vacation.User = &user

	return &vacation, err
}

func (u *vacationsUsecase) GetAll(ctx context.Context) ([]models.Vacation, error) {
	vacationsDB, err := u.VacationRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return u.joinVacationsWithUser(ctx, vacationsDB...)
}

func (u *vacationsUsecase) joinVacationsWithUser(ctx context.Context, vacationsDB ...models.VacationDB) ([]models.Vacation, error) {
	var (
		txID      = transactionID.FromContext(ctx)
		vacations = make([]models.Vacation, 0, len(vacationsDB))
		users     = make(map[string]models.User, len(vacationsDB)*2)
		err       error
	)

	for _, vac := range vacationsDB {
		v := copyToVacation(vac)
		user, ok := users[vac.UserID]
		if !ok {
			user, err = u.userRepo.GetUserByID(ctx, vac.UserID)
			if err != nil {
				u.log.Warnf(txID, "skipping vacation due to invalid userID=%s", vac.UserID)
				continue
			}
			users[vac.UserID] = user
		}
		v.User = &user

		if vac.StatusChangerID == "" {
			vacations = append(vacations, v)
			continue
		}

		admin, ok := users[vac.StatusChangerID]
		if !ok {
			admin, err = u.userRepo.GetUserByID(ctx, vac.StatusChangerID)
			if err != nil {
				u.log.Warnf(txID, "skipping vacation due to invalid statusChangerID=%s", vac.StatusChangerID)
				continue
			}
			users[vac.StatusChangerID] = admin
		}
		v.StatusChanger = &admin
		vacations = append(vacations, v)
	}
	return vacations, nil
}

func (u *vacationsUsecase) GetPending(ctx context.Context) ([]models.Vacation, error) {
	vacationsDB, err := u.VacationRepository.GetPending(ctx)
	if err != nil {
		return nil, err
	}

	return u.joinVacationsWithUser(ctx, vacationsDB...)
}
func (u *vacationsUsecase) GetForUser(ctx context.Context, userID string) ([]models.Vacation, error) {
	vacationsDB, err := u.VacationRepository.GetForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return u.joinVacationsWithUser(ctx, vacationsDB...)
}

func copyToVacation(v models.VacationDB) models.Vacation {
	return models.Vacation{
		ID:          v.ID,
		Number:      v.Number,
		StartDate:   v.StartDate,
		EndDate:     v.EndDate,
		Status:      v.Status,
		UpdateTime:  v.UpdateTime,
		WasApproved: v.WasApproved,
	}
}

func isTimeIntersected(existedStart, existedEnd, newStart, newEnd time.Time) bool {
	input := spaniel.Spans{
		spaniel.New(existedStart, existedEnd),
		spaniel.New(newStart, newEnd),
	}

	inter := input.Intersection()
	return len(inter) > 0
}
