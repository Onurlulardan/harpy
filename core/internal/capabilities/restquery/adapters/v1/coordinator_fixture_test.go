package v1query

import (
	"context"
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

type coordinatorAuthorizer struct {
	t      *testing.T
	events *[]string
}

func (port coordinatorAuthorizer) ResolveAndAuthorizeV1(
	_ context.Context, request api.QueryRequest,
) (resolved.AuthorizedPlan, error) {
	*port.events = append(*port.events, "authorize")
	return coordinatorPlan(port.t, request), nil
}

type coordinatorEvidence struct {
	events  *[]string
	retired []contractretirement.Key
	admit   int
	hit     int
}

func (port *coordinatorEvidence) AdmitManifest(
	_ context.Context, batch contractretirement.Batch,
) (contractretirement.Decision, error) {
	*port.events = append(*port.events, "admit")
	port.admit++
	return contractretirement.SealDecision(batch, port.retired)
}

func (port *coordinatorEvidence) RecordHits(
	_ context.Context, batch contractretirement.Batch,
) (contractretirement.Decision, error) {
	*port.events = append(*port.events, "hit")
	port.hit++
	return contractretirement.SealDecision(batch, port.retired)
}

func coordinatorPlan(t *testing.T, request api.QueryRequest) resolved.AuthorizedPlan {
	t.Helper()
	spec, err := resolver.Resolve(request, coordinatorCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	reference, _ := spec.Reference()
	path, _ := spec.RootPath()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("v1-coordinator-requirement")), reference, path,
	)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		uuid.MustParse("30000000-0000-4000-8000-000000000002"),
		sha256.Sum256([]byte("v1-coordinator-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func coordinatorCatalog(t *testing.T) api.CatalogSet {
	t.Helper()
	reference, err := api.NewCatalogReference(
		"forms", "10000000-0000-4000-8000-000000000001",
		"10000000-0000-4000-8000-000000000002",
		mustCoordinatorHash(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	id := coordinatorField(t, "20000000-0000-4000-8000-000000000001", "id", true)
	status := coordinatorField(
		t, "20000000-0000-4000-8000-000000000002", "status", false,
	)
	createdAt := coordinatorField(
		t, "20000000-0000-4000-8000-000000000003", "created_at", false,
	)
	idValue, _ := id.ID()
	statusValue, _ := status.ID()
	table, _ := sqlidentifier.MintDynamic("frm_v1_coordinator")
	storage, _ := api.NewResourceStorage(table)
	capabilities, _ := api.NewCatalogCapabilities(true)
	catalog, err := api.NewFieldCatalog(api.FieldCatalogInput{
		Reference: reference, Storage: storage, Capabilities: capabilities,
		Fields:        []api.FieldDefinition{id, status, createdAt},
		Relations:     []api.RelationDefinition{},
		DefaultSelect: []uuid.UUID{idValue, statusValue}, StableID: idValue,
		SearchFields: []uuid.UUID{statusValue},
	})
	if err != nil {
		t.Fatal(err)
	}
	set, err := api.NewCatalogSet(reference, []api.FieldCatalog{catalog})
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func coordinatorField(
	t *testing.T, id string, name string, uuidField bool,
) api.FieldDefinition {
	t.Helper()
	var fieldType api.FieldType
	if uuidField {
		fieldType, _ = api.NewUUIDFieldType()
	} else {
		profile, _ := api.NewTextProfile(
			api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
		)
		fieldType, _ = api.NewTextFieldType(profile)
	}
	column, _ := sqlidentifier.MintDynamic("fld_" + name)
	var folded *sqlidentifier.Token
	if !uuidField {
		value, _ := sqlidentifier.MintDynamic("fld_" + name + "_folded")
		folded = &value
	}
	storage, _ := api.NewFieldStorage(column, folded, nil)
	canonicalBytes := 256
	if uuidField {
		canonicalBytes = 36
	}
	input := api.FieldCapabilitiesInput{
		Selectable: true, Sortable: true, CursorCanonicalBytes: canonicalBytes,
		Filters: []api.FilterCapabilityInput{{Operator: api.OperatorEq, CountSafe: true}},
	}
	if !uuidField {
		input.Searchable = true
		input.Filters = append(input.Filters, api.FilterCapabilityInput{
			Operator: api.OperatorContains, CountSafe: true,
		})
	}
	capabilities, _ := api.NewFieldCapabilities(fieldType, input)
	field, err := api.NewFieldDefinition(id, name, fieldType, storage, capabilities)
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func mustCoordinatorHash(t *testing.T) api.CatalogHash {
	t.Helper()
	hash, err := api.NewCatalogHash(strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func coordinatorKey(t *testing.T, aliasID string) contractretirement.Key {
	t.Helper()
	key, err := NewRetirementKey(testManifest().ManifestID, aliasID)
	if err != nil {
		t.Fatal(err)
	}
	value, _ := key.Value()
	parsed, err := contractretirement.ParseKey(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
