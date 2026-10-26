package restoreexecution

import (
	"errors"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

var (
	ErrClaim      = errors.New("restore execution claim failed")
	ErrTransition = errors.New("restore execution transition failed")
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	codePattern   = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)
	canonicalUUID = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
)

type Service struct{ store ports.RestoreExecutionStore }

func New(store ports.RestoreExecutionStore) *Service {
	if store == nil {
		return nil
	}
	return &Service{store: store}
}

func changed(err error, value bool) error {
	if err != nil || !value {
		return ErrTransition
	}
	return nil
}
