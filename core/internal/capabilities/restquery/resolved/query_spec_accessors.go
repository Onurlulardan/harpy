package resolved

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (spec QuerySpec) Mode() (api.QueryMode, error) {
	if err := spec.Validate(); err != nil {
		return 0, err
	}
	return spec.mode, nil
}

func (spec QuerySpec) Catalogs() (api.CatalogSet, error) {
	if err := spec.Validate(); err != nil {
		return api.CatalogSet{}, err
	}
	return spec.catalogs, nil
}

func (spec QuerySpec) Reference() (api.CatalogReference, error) {
	if err := spec.Validate(); err != nil {
		return api.CatalogReference{}, err
	}
	return spec.reference, nil
}

func (spec QuerySpec) Storage() (api.ResourceStorage, error) {
	if err := spec.Validate(); err != nil {
		return api.ResourceStorage{}, err
	}
	return spec.storage, nil
}

func (spec QuerySpec) StableID() (api.FieldDefinition, error) {
	if err := spec.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return spec.stableID, nil
}

func (spec QuerySpec) RootPath() (RequestPath, error) {
	if err := spec.Validate(); err != nil {
		return RequestPath{}, err
	}
	return spec.rootPath, nil
}

func (spec QuerySpec) Record() (RecordQuery, bool, error) {
	if err := spec.Validate(); err != nil {
		return RecordQuery{}, false, err
	}
	if spec.record == nil {
		return RecordQuery{}, false, nil
	}
	return *spec.record, true, nil
}

func (spec QuerySpec) Aggregate() (AggregateQuery, bool, error) {
	if err := spec.Validate(); err != nil {
		return AggregateQuery{}, false, err
	}
	if spec.aggregate == nil {
		return AggregateQuery{}, false, nil
	}
	return *spec.aggregate, true, nil
}

func (spec QuerySpec) IdentityHash() ([sha256.Size]byte, error) {
	if err := spec.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return spec.identityHash, nil
}

func (spec QuerySpec) IdentitySize() (int, error) {
	if err := spec.Validate(); err != nil {
		return 0, err
	}
	return spec.identitySize, nil
}
