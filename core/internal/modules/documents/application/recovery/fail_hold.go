package recovery

import (
	"context"
	"regexp"

	"github.com/google/uuid"
)

var failureCodePattern = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)

func (service *Service) FailHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
	code string,
) error {
	if !service.usable(ctx) || recoverySetID == uuid.Nil ||
		!failureCodePattern.MatchString(code) {
		return ErrInvalid
	}
	changed, err := service.store.FailHold(ctx, recoverySetID, code)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}
