package systemresources

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

const (
	testTenantID  = "10000000-0000-4000-8000-000000000001"
	testActorID   = "20000000-0000-4000-8000-000000000002"
	testRecordID  = "018f47f4-7b1b-7cc0-8000-000000000010"
	testVersionID = "018f47f4-7b1b-7cc0-8000-000000000011"
	testFieldID   = "a1000000-0000-4000-8000-000000000001"
)

type testFixture struct {
	catalog   publicresource.FieldCatalog
	document  formpresentation.Document
	read      publicresource.ReadResult
	mutation  publicresource.MutationResult
	query     publicresource.QueryResult
	deleted   publicresource.DeleteResult
	admission Admission
	values    publicresource.FieldValues
}

func newTestFixture(t *testing.T) testFixture {
	t.Helper()
	field, err := formpresentation.NewField(
		testFieldID, fieldregistry.TypeText, fieldregistry.ValueTextNFCV1,
		formpresentation.FieldEditable,
	)
	if err != nil {
		t.Fatal(err)
	}
	fields, err := formpresentation.NewFieldCatalog([]formpresentation.Field{field})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := publicresource.NewFieldCatalog(
		publicresource.Users, mustID(t, publicresource.UsersCatalogResourceID),
		mustID(t, testVersionID), sha256.Sum256([]byte("query")), fields,
		[]publicresource.FieldDescriptor{mustDescriptor(t, testFieldID, "value")},
	)
	if err != nil {
		t.Fatal(err)
	}
	values, err := publicresource.NewFieldValues(map[publicresource.ID]json.RawMessage{
		mustID(t, testFieldID): json.RawMessage(`"value"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	record, _ := publicresource.NewRecord(publicresource.Users, mustID(t, testRecordID), 2, values)
	read, _ := publicresource.NewReadResult(record, catalog)
	receipt, _ := publicresource.NewMutationReceipt(
		mustID(t, "018f47f4-7b1b-7cc0-8000-000000000012"),
		sha256.Sum256([]byte("request")), 2, false,
	)
	mutation, _ := publicresource.NewMutationResult(read, receipt)
	query, _ := publicresource.NewQueryResult(
		[]json.RawMessage{json.RawMessage(`{}`)}, false, "", nil, catalog,
	)
	deleted, _ := publicresource.NewDeleteResult(publicresource.Users, catalog, receipt)
	actor, _ := publicresource.NewActorEvidence(
		mustID(t, testTenantID), mustID(t, testActorID), 1, 1,
	)
	admission, _ := NewAdmission(mustID(t, testTenantID), actor, 3, 3, 3)
	return testFixture{
		catalog: catalog, document: testDocument(), read: read, mutation: mutation,
		query: query, deleted: deleted, admission: admission, values: values,
	}
}

func mustID(t *testing.T, value string) publicresource.ID {
	t.Helper()
	id, err := publicresource.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func mustDescriptor(
	t *testing.T, id, name string,
) publicresource.FieldDescriptor {
	t.Helper()
	descriptor, err := publicresource.NewFieldDescriptor(mustID(t, id), name)
	if err != nil {
		t.Fatal(err)
	}
	return descriptor
}

func testDocument() formpresentation.Document {
	return formpresentation.Document{
		RegistryContract: formpresentation.RegistryContractVersion,
		DefaultLocale:    "en", Locales: []string{"en"},
		Surfaces: []formpresentation.Surface{{
			ID: "018f47f4-7b1b-7cc0-8000-000000000020", Kind: "record_update",
			Translations: []formpresentation.Translation{{Locale: "en", Label: "Update"}},
			Root: formpresentation.Node{
				ID: "018f47f4-7b1b-7cc0-8000-000000000021", Kind: "container",
				Variant: "stack", Span: 1, Translations: []formpresentation.Translation{},
				Children: []formpresentation.Node{{
					ID: "018f47f4-7b1b-7cc0-8000-000000000022", Kind: "field",
					FieldID: testFieldID, Widget: "text", Interaction: "auto", Span: 1,
					Translations: []formpresentation.Translation{{Locale: "en", Label: "Value"}},
				}},
			},
		}},
	}
}

func testQueryInput(t *testing.T) publicresource.QueryInput {
	t.Helper()
	page, _ := restquery.NewRequestPage(10, nil)
	request, _ := restquery.NewRecordQueryRequest(restquery.RecordQueryRequestInput{
		Page: page, Count: restquery.CountModeNone,
	})
	input, err := publicresource.NewQueryInput(publicresource.Users, request)
	if err != nil {
		t.Fatal(err)
	}
	return input
}
