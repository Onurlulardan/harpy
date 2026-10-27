package legalhold

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const (
	OperationPlace   = "documents.legal_hold.place"
	OperationRelease = "documents.legal_hold.release"
)

type Application struct {
	binder    StoreBinder
	authority SourceAuthority
	ids       foundation.IDSource
	clock     foundation.Clock
}

func NewApplication(binder StoreBinder, authority SourceAuthority,
	ids foundation.IDSource, clock foundation.Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authority, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents legal hold dependencies are invalid")
		}
	}
	return &Application{binder: binder, authority: authority, ids: ids, clock: clock}, nil
}

func (application *Application) identifiers() (domain.ID, domain.ID, domain.ID, error) {
	values := make([]domain.ID, 3)
	seen := map[domain.ID]struct{}{}
	for index := range values {
		value, err := application.ids.NewUUIDv7()
		_, duplicate := seen[value]
		if err != nil || !value.UUIDv7() || duplicate {
			return domain.ID{}, domain.ID{}, domain.ID{}, ErrUnavailable
		}
		values[index], seen[value] = value, struct{}{}
	}
	return values[0], values[1], values[2], nil
}
