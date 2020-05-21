package vacation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dimitriy14/staff-manager/models"
	"github.com/Dimitriy14/staff-manager/repository"
	"github.com/Dimitriy14/staff-manager/util"
	"github.com/senseyeio/spaniel"

	"github.com/google/uuid"
)

const day = time.Hour * time.Duration(24)

type VacationsUsecase interface {
	Save(ctx context.Context, vacation models.VacationDB) (*models.Vacation, error)
	UpdateVacationStatus(ctx context.Context, vacationID uuid.UUID, status models.VacationStatus) (*models.Vacation, error)
	GetAll(ctx context.Context) ([]models.VacationDB, error)
	GetPending(ctx context.Context) ([]models.VacationDB, error)
	GetForUser(ctx context.Context, userID string) ([]models.VacationDB, error)
	GetByID(ctx context.Context, vacationID uuid.UUID) (*models.Vacation, error)
}

func NewVacationUseCase(vacationRepo repository.VacationRepository,
	userRepo repository.UserRepository,
	recentChangesRepo repository.RecentActionRepository) *vacationsUsecase {
	return &vacationsUsecase{
		VacationRepository: vacationRepo,
		userRepo:           userRepo,
		recentChangesRepo:  recentChangesRepo,
	}
}

type vacationsUsecase struct {
	repository.VacationRepository
	userRepo          repository.UserRepository
	recentChangesRepo repository.RecentActionRepository
}

func (u *vacationsUsecase) Save(ctx context.Context, vacation models.VacationDB) (*models.Vacation, error) {
	actualVacations, err := u.VacationRepository.GetPendingForUser(ctx, vacation.UserID)
	if err != nil {
		return nil, err
	}

	for _, actualVacation := range actualVacations {
		if isTimeIntersected(
			actualVacation.StartDate, actualVacation.EndDate,
			vacation.StartDate, vacation.EndDate,
		) {
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
		Title:      "Vacation",
		IncidentID: vacation.ID,
		Type:       models.VacationRequest,
		UserName:   vacation.UserFullName,
		UserID:     user.ID.String(),
		OwnerID:    user.ID.String(),
		ChangeTime: vacation.UpdateTime,
	})

	return &vac, err
}

func (u *vacationsUsecase) UpdateVacationStatus(ctx context.Context, vacationID uuid.UUID, status models.VacationStatus) (*models.Vacation, error) {
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
	oldVacation.UpdateTime = time.Now().UTC()
	vac := copyToVacation(*oldVacation)
	vac.User = &user
	vac.StatusChanger = &user

	err = u.VacationRepository.Update(ctx, *oldVacation)
	if err != nil {
		return nil, err
	}

	err = u.recentChangesRepo.Save(models.RecentChanges{
		ID:            uuid.New(),
		Title:         string(status),
		IncidentID:    oldVacation.ID,
		Type:          models.VacationStatusChange,
		UserName:      fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		UserID:        user.ID.String(),
		OwnerID:       user.ID.String(),
		UpdatedByName: oldVacation.StatusChangerFullName,
		UpdatedByID:   oldVacation.StatusChangerID,
		ChangeTime:    oldVacation.UpdateTime,
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
