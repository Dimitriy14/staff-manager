package recent

import (
	"github.com/Dimitriy14/staff-manager/db"
	"github.com/Dimitriy14/staff-manager/models"

	"github.com/pkg/errors"
)

func NewRecentActionRepo(client *db.Client) *recentActionRepo {
	return &recentActionRepo{client}
}

type recentActionRepo struct {
	*db.Client
}

func (r *recentActionRepo) Save(action models.RecentChanges) error {
	errs := r.Session.Save(&action).GetErrors()
	if len(errs) > 1 {
		return errors.Wrap(concatErrors(errs...), "saving action error")
	}
	return nil
}

func (r *recentActionRepo) GetUserChanges(userID string) ([]models.RecentChanges, error) {
	actions := make([]models.RecentChanges, 0, 0)
	errs := r.Session.Where("UserID = ? OR OwnerID = ?", userID, userID).Order("ChangeTime desc").Find(&actions).GetErrors()
	if len(errs) > 1 {
		return nil, errors.Wrap(concatErrors(errs...), "getting action error")
	}
	return actions, nil
}

func concatErrors(errs ...error) error {
	var e string
	for _, err := range errs {
		e += err.Error()
	}
	return errors.New(e)
}
