package systemresources

import (
	"crypto/sha256"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestPresentationBindsExactCatalogAndDocumentEvidence(t *testing.T) {
	fieldID := "a1000000-0000-4000-8000-000000000001"
	field, _ := formpresentation.NewField(
		fieldID, fieldregistry.TypeText, fieldregistry.ValueTextNFCV1,
		formpresentation.FieldEditable,
	)
	fields, _ := formpresentation.NewFieldCatalog([]formpresentation.Field{field})
	resourceID, _ := publicresource.ParseID(publicresource.UsersCatalogResourceID)
	versionID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000011")
	catalog, _ := publicresource.NewFieldCatalog(
		publicresource.Users, resourceID, versionID, sha256.Sum256([]byte("query")), fields,
		[]publicresource.FieldDescriptor{mustDescriptor(t, fieldID, "value")},
	)
	document := formpresentation.Document{
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
					FieldID: fieldID, Widget: "text", Interaction: "auto", Span: 1,
					Translations: []formpresentation.Translation{{Locale: "en", Label: "Value"}},
				}},
			},
		}},
	}
	presentation, err := NewPresentation(catalog, document)
	catalogSHA, documentSHA, evidenceErr := presentation.Evidence()
	if err != nil || evidenceErr != nil || catalogSHA == [32]byte{} ||
		documentSHA == [32]byte{} || presentation.Validate() != nil {
		t.Fatalf("presentation error=%v evidence=%v", err, evidenceErr)
	}
}
