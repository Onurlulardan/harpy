package link

import (
	"fmt"

	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type FormRecordRegistry struct {
	binder StoreBinder
	ids    foundation.IDSource
	clock  foundation.Clock
}

func NewFormRecordRegistry(
	binder StoreBinder,
	ids foundation.IDSource,
	clock foundation.Clock,
) (*FormRecordRegistry, error) {
	if foundation.NilPort(binder) || foundation.NilPort(ids) ||
		foundation.NilPort(clock) {
		return nil, fmt.Errorf("documents form-record link dependencies are invalid")
	}
	return &FormRecordRegistry{binder: binder, ids: ids, clock: clock}, nil
}

var _ formrecordlink.Registry = (*FormRecordRegistry)(nil)
var _ formrecordlink.AdoptionRegistry = (*FormRecordRegistry)(nil)
