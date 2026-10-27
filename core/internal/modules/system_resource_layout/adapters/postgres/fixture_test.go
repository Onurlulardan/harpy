package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

var (
	adapterRevisionID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	adapterNow        = time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
)

const adapterField = `{"field_id":"33333333-3333-4333-8333-333333333333",` +
	`"interaction":"auto","order":0,"translations":[],"widget":"auto"}`

const adapterSection = `{"fields":[` + adapterField + `],` +
	`"id":"22222222-2222-4222-8222-222222222222","order":0,"translations":[]}`

const adapterDefinition = `{"contract":"harpy/system-resource-layout/v2",` +
	`"surfaces":[` +
	`{"kind":"record_create","sections":[` + adapterSection + `],"translations":[]},` +
	`{"kind":"record_update","sections":[` + adapterSection + `],"translations":[]},` +
	`{"kind":"record_detail","sections":[` + adapterSection + `],"translations":[]},` +
	`{"kind":"record_list","sections":[` + adapterSection + `],"translations":[]},` +
	`{"kind":"record_filter","sections":[` + adapterSection + `],"translations":[]}]}`

const adapterLegacyDefinition = `{"contract":"harpy/system-resource-layout/v1",` +
	`"sections":[{"fields":[{"field_id":"33333333-3333-4333-8333-333333333333",` +
	`"order":0,"translations":[],"widget":"auto"}],` +
	`"id":"22222222-2222-4222-8222-222222222222","order":0,"translations":[]}]}`

type fixtureRow struct {
	document []byte
	boolean  bool
	err      error
}

func (row fixtureRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 1 {
		return layoutUnavailable()
	}
	switch destination := destinations[0].(type) {
	case *[]byte:
		*destination = append([]byte(nil), row.document...)
	case *bool:
		*destination = row.boolean
	default:
		return layoutUnavailable()
	}
	return nil
}

func fixtureRevisionJSON(
	t *testing.T,
	state layoutapi.RevisionState,
) []byte {
	t.Helper()
	digest := sha256.Sum256([]byte(adapterDefinition))
	wire := revisionWire{
		ID: adapterRevisionID.String(), ResourceKind: "user", Version: 1,
		State: string(state), Definition: json.RawMessage(adapterDefinition),
		DefinitionSHA256: hex.EncodeToString(digest[:]),
		CreatedAt:        adapterNow, UpdatedAt: adapterNow, RowRevision: 1,
	}
	if state == layoutapi.RevisionPublished {
		wire.PublishedAt = &wire.UpdatedAt
		wire.RowRevision = 2
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func fixturePublicationJSON(t *testing.T, complete bool) []byte {
	t.Helper()
	wire := publicationWire{
		Revision:   fixtureRevisionJSON(t, layoutapi.RevisionPublished),
		Generation: 1,
	}
	if complete {
		sourceSHA := sha256.Sum256([]byte(adapterDefinition))
		catalogJSON := `{"contract":"forms-presentation-field-catalog.v1","fields":[]}`
		catalogSHA := sha256.Sum256([]byte(catalogJSON))
		presentationJSON := `{"registry_contract":"forms-presentation-registry.v1"}`
		presentationSHA := sha256.Sum256([]byte(presentationJSON))
		wire.SourceDefinitionSHA256 = hex.EncodeToString(sourceSHA[:])
		wire.CompilerVersion = "system-resource-presentation-compiler.v2"
		wire.CatalogContract = "forms-presentation-field-catalog.v1"
		wire.CatalogJSON = catalogJSON
		wire.CatalogSHA256 = hex.EncodeToString(catalogSHA[:])
		wire.PresentationContract = "forms-presentation.v1"
		wire.PresentationJSON = presentationJSON
		wire.PresentationSHA256 = hex.EncodeToString(presentationSHA[:])
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func fixtureLayoutStore(rowValue row) *layoutStore {
	return &layoutStore{
		queryRow: func(_ context.Context, _ string, _ ...any) row {
			return rowValue
		},
		commandAllowed: func(context.Context) bool { return true },
		queryAllowed:   func(context.Context) bool { return true },
	}
}
