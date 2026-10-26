package documents

import (
	"fmt"

	linkpostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/link"
	formrecordlink "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
)

type FormRecordLinkRegistry interface {
	formrecordlink.Registry
	formrecordlink.AdoptionRegistry
}

func NewFormRecordLinkRegistry() (FormRecordLinkRegistry, error) {
	registry, err := linkapp.NewFormRecordRegistry(
		linkpostgres.NewBinder(), platformIDs{}, platformClock{},
	)
	if err != nil {
		return nil, fmt.Errorf("build Documents form-record link registry: %w", err)
	}
	return registry, nil
}
