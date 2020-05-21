package vacation

import (
	"context"

	"github.com/Dimitriy14/staff-manager/db"
	"github.com/Dimitriy14/staff-manager/models"

	"github.com/pkg/errors"
)

func NewVacationRepo(client *db.Client) *vacationRepo {
	return &vacationRepo{client}
}

type vacationRepo struct {
	*db.Client
}

func (r *vacationRepo) Save(_ context.Context, vacation models.VacationDB) (*models.VacationDB, error) {
	errs := r.Session.Save(&vacation).GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "saving vacation error")
	}
	return &vacation, nil
}

func (r *vacationRepo) Update(_ context.Context, vacation models.VacationDB) error {
	errs := r.Session.Update(&vacation).GetErrors()
	if len(errs) > 1 {
		return errors.Wrap(concatErrors(errs...), "updating vacation error")
	}
	return nil
}

func (r *vacationRepo) GetAll(_ context.Context) ([]models.VacationDB, error) {
	vacations := make([]models.VacationDB, 0)
	errs := r.Session.Where("status in (?, ?, ?)", models.Pending, models.Approved, models.Rejected).
		Order("start_date").
		Find(&vacations).
		GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "getting all vacation error")
	}
	return vacations, nil
}

func (r *vacationRepo) GetPending(_ context.Context) ([]models.VacationDB, error) {
	vacations := make([]models.VacationDB, 0)
	errs := r.Session.Where("status = ?", models.Pending).
		Order("start_date").
		Find(&vacations).
		GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "getting pending vacation error")
	}
	return vacations, nil
}

func (r *vacationRepo) GetForUser(_ context.Context, userID string) ([]models.VacationDB, error) {
	vacations := make([]models.VacationDB, 0)
	errs := r.Session.Where("user_id = ?", userID).
		Order("start_date").
		Find(&vacations).
		GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "getting user vacation error")
	}
	return vacations, nil
}

func (r *vacationRepo) GetPendingForUser(_ context.Context, userID string) ([]models.VacationDB, error) {
	vacations := make([]models.VacationDB, 0)
	errs := r.Session.Where("user_id = ? AND status = ?", userID, models.Pending).
		Order("start_date").
		Find(&vacations).
		GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "getting user vacation error")
	}
	return vacations, nil
}

func (r *vacationRepo) GetByID(ctx context.Context, vacationID string) (*models.VacationDB, error) {
	var vacation = new(models.VacationDB)
	errs := r.Session.Where("id = ?", vacationID).
		Find(vacation).
		GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "getting user vacation error")
	}
	return vacation, nil
}

func concatErrors(errs ...error) error {
	var e string
	for _, err := range errs {
		e += err.Error()
	}
	return errors.New(e)
}
