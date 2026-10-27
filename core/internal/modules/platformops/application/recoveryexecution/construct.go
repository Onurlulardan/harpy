package recoveryexecution

import (
	"errors"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	"github.com/google/uuid"
)

var (
	ErrInvalid     = errors.New("recovery execution input is invalid")
	ErrConflict    = errors.New("recovery execution state conflicts")
	ErrUnavailable = errors.New("recovery execution is unavailable")
	digestPattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	codePattern    = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)
)

type Service struct{ store ports.RecoveryExecutionStore }

func New(store ports.RecoveryExecutionStore) (*Service, error) {
	if store == nil {
		return nil, ErrUnavailable
	}
	return &Service{store: store}, nil
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.String() == value
}

func validUUIDv7(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.String() == value && parsed.Version() == 7
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }

func validIdentity(value ports.RecoveryExecutionIdentity) bool {
	return validUUID(value.RecoverySetID) && validUUIDv7(value.AttemptID) &&
		validUUIDv7(value.LeaseToken)
}
