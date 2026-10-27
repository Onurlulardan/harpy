package storagecatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

var queryNamespace = uuid.MustParse("e7d1689a-8ef2-4b10-99ac-3fcda44bc76e")

func newCatalog(spec queryResourceSpec) (workquery.ResolvedCatalog, error) {
	hash, err := catalogHash(spec)
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	reference, err := restapi.NewCatalogReference(workquery.Owner,
		queryID(spec.name, "resource").String(),
		queryID(spec.name, "published:v1").String(), hash)
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	table, err := sqlidentifier.MintStatic(spec.view)
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	storage, err := restapi.NewResourceStorage(table)
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	fields, defaults, searches, stable, err := buildFields(spec)
	if err != nil {
		return workquery.ResolvedCatalog{}, err
	}
	capabilities, err := restapi.NewCatalogCapabilities(false)
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	root, err := restapi.NewFieldCatalog(restapi.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields: fields, Relations: []restapi.RelationDefinition{},
		DefaultSelect: defaults, StableID: stable, SearchFields: searches,
	})
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	set, err := restapi.NewCatalogSet(reference, []restapi.FieldCatalog{root})
	if err != nil {
		return workquery.ResolvedCatalog{}, workquery.ErrUnavailable
	}
	return workquery.NewResolvedCatalog(spec.name, set)
}

func buildFields(spec queryResourceSpec) (
	[]restapi.FieldDefinition, []uuid.UUID, []uuid.UUID, uuid.UUID, error,
) {
	fields := make([]restapi.FieldDefinition, len(spec.fields))
	defaults := make([]uuid.UUID, 0, len(spec.fields))
	searches := make([]uuid.UUID, 0, 2)
	stable := uuid.Nil
	for index, fieldSpec := range spec.fields {
		field, err := buildField(spec.name, fieldSpec)
		if err != nil {
			return nil, nil, nil, uuid.Nil, err
		}
		identifier, err := field.ID()
		if err != nil {
			return nil, nil, nil, uuid.Nil, workquery.ErrUnavailable
		}
		fields[index] = field
		if fieldSpec.selectable {
			defaults = append(defaults, identifier)
		}
		if fieldSpec.searchable {
			searches = append(searches, identifier)
		}
		if fieldSpec.name == spec.stable {
			stable = identifier
		}
	}
	if stable == uuid.Nil {
		return nil, nil, nil, uuid.Nil, workquery.ErrUnavailable
	}
	return fields, defaults, searches, stable, nil
}

func queryID(resource, member string) uuid.UUID {
	return uuid.NewSHA1(queryNamespace, []byte("work-rest-query:v1:"+resource+":"+member))
}

func catalogHash(spec queryResourceSpec) (restapi.CatalogHash, error) {
	fields := make([]map[string]any, len(spec.fields))
	for index, field := range spec.fields {
		fields[index] = map[string]any{"name": field.name, "column": field.column,
			"folded": field.folded, "kind": field.kind,
			"selectable": field.selectable, "sortable": field.sortable,
			"searchable": field.searchable, "nullable": field.nullable}
	}
	document := map[string]any{"contract": "work-rest-query-catalog.v1",
		"resource": spec.name, "view": spec.view, "operation": spec.operation,
		"stable": spec.stable, "fields": fields}
	canonical, err := json.Marshal(document)
	if err != nil {
		return restapi.CatalogHash{}, err
	}
	digest := sha256.Sum256(canonical)
	return restapi.NewCatalogHash(hex.EncodeToString(digest[:]))
}
