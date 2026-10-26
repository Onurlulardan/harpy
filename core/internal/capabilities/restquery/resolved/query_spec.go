package resolved

import (
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const MaximumIdentityBytes = 64 * 1024

type QuerySpec struct {
	mode         api.QueryMode
	catalogs     api.CatalogSet
	reference    api.CatalogReference
	storage      api.ResourceStorage
	stableID     api.FieldDefinition
	rootPath     RequestPath
	record       *RecordQuery
	aggregate    *AggregateQuery
	identityHash [sha256.Size]byte
	stateSeal    [sha256.Size]byte
	catalogSeal  [sha256.Size]byte
	identitySize int
	stateSize    int
	validated    bool
}

func NewRecordQuerySpec(
	catalogs api.CatalogSet,
	reference api.CatalogReference,
	storage api.ResourceStorage,
	stableID api.FieldDefinition,
	rootPath RequestPath,
	record RecordQuery,
) (QuerySpec, error) {
	owned := record
	spec := QuerySpec{
		mode: api.QueryModeRecord, catalogs: catalogs, reference: reference, storage: storage,
		stableID: stableID, rootPath: rootPath, record: &owned, validated: true,
	}
	return sealQuerySpec(spec)
}

func NewAggregateQuerySpec(
	catalogs api.CatalogSet,
	reference api.CatalogReference,
	storage api.ResourceStorage,
	stableID api.FieldDefinition,
	rootPath RequestPath,
	aggregate AggregateQuery,
) (QuerySpec, error) {
	owned := aggregate
	spec := QuerySpec{
		mode: api.QueryModeAggregate, catalogs: catalogs,
		reference: reference, storage: storage,
		stableID: stableID, rootPath: rootPath, aggregate: &owned, validated: true,
	}
	return sealQuerySpec(spec)
}

func sealQuerySpec(spec QuerySpec) (QuerySpec, error) {
	if err := spec.validateShape(); err != nil {
		return QuerySpec{}, err
	}
	identity, seal, identitySize, stateSize, err := queryStateFingerprints(spec)
	if err != nil {
		return QuerySpec{}, err
	}
	spec.identityHash = identity
	spec.stateSeal = seal
	spec.catalogSeal, err = catalogSetFingerprint(spec.catalogs)
	if err != nil {
		return QuerySpec{}, err
	}
	spec.identitySize = identitySize
	spec.stateSize = stateSize
	if err := spec.Validate(); err != nil {
		return QuerySpec{}, err
	}
	return spec, nil
}

func (spec QuerySpec) Validate() error {
	if !spec.validated || spec.validateShape() != nil ||
		spec.identityHash == [sha256.Size]byte{} ||
		spec.stateSeal == [sha256.Size]byte{} || spec.catalogSeal == [sha256.Size]byte{} ||
		spec.identitySize < 1 || spec.identitySize > MaximumIdentityBytes ||
		spec.stateSize < 1 || spec.stateSize > MaximumIdentityBytes {
		return fmt.Errorf("resolved REST Query spec is invalid")
	}
	identity, seal, identitySize, stateSize, err := queryStateFingerprints(spec)
	catalogSeal, catalogErr := catalogSetFingerprint(spec.catalogs)
	if err != nil || identity != spec.identityHash || seal != spec.stateSeal ||
		identitySize != spec.identitySize || stateSize != spec.stateSize ||
		catalogErr != nil || catalogSeal != spec.catalogSeal {
		return fmt.Errorf("resolved REST Query spec is invalid")
	}
	return nil
}

func (spec QuerySpec) validateShape() error {
	if spec.catalogs.Validate() != nil || spec.reference.Validate() != nil ||
		spec.storage.Validate() != nil ||
		spec.stableID.Validate() != nil || spec.rootPath.Validate() != nil ||
		!validStableField(spec.stableID) || !specRootMatchesCatalogs(spec) ||
		!specMembersMatchCatalogs(spec) || !specPathsCanonical(spec) {
		return fmt.Errorf("resolved REST Query spec shape is invalid")
	}
	if spec.mode == api.QueryModeRecord {
		if spec.record == nil || spec.aggregate != nil || spec.record.Validate() != nil ||
			!validRecordSpec(*spec.record, spec.reference, spec.stableID) {
			return fmt.Errorf("resolved REST Query spec shape is invalid")
		}
		return nil
	}
	if spec.mode == api.QueryModeAggregate {
		if spec.aggregate == nil || spec.record != nil || spec.aggregate.Validate() != nil {
			return fmt.Errorf("resolved REST Query spec shape is invalid")
		}
		return nil
	}
	return fmt.Errorf("resolved REST Query spec shape is invalid")
}

func (QuerySpec) String() string                      { return valueRedaction }
func (QuerySpec) GoString() string                    { return valueRedaction }
func (spec QuerySpec) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (QuerySpec) MarshalJSON() ([]byte, error)        { return rejectDirectJSONMarshal() }
