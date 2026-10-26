package presentation

import (
	"bytes"
	"errors"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
)

func TestDocumentAcceptsStableIAMFieldAndStrictWidget(t *testing.T) {
	catalog := testCatalog(t, FieldEditable)
	document := testDocument("text", "auto")
	encoded, digest, err := EncodeDocument(document, catalog)
	if err != nil || digest == [32]byte{} {
		t.Fatalf("encode document: %v", err)
	}
	decoded, err := DecodeDocument(encoded, catalog)
	if err != nil || decoded.Surfaces[0].Root.Children[0].FieldID != testTextFieldID {
		t.Fatalf("decode document: %+v, %v", decoded, err)
	}
	if _, err := DecodeDocument(append(encoded, ' '), catalog); !errors.Is(err, ErrInvalid) {
		t.Fatalf("non-canonical document error = %v", err)
	}
}

func TestDocumentFailsClosedOnCatalogAndWidgetDrift(t *testing.T) {
	for name, test := range map[string]struct {
		catalog  FieldCatalog
		document Document
	}{
		"unknown_widget": {testCatalog(t, FieldEditable), testDocument("mystery", "auto")},
		"wrong_widget":   {testCatalog(t, FieldEditable), testDocument("toggle", "auto")},
		"readonly_auto":  {testCatalog(t, FieldReadOnly), testDocument("text", "auto")},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := EncodeDocument(test.document, test.catalog); !errors.Is(err, ErrInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	readOnly := testCatalog(t, FieldReadOnly)
	if _, _, err := EncodeDocument(testDocument("text", "read_only"), readOnly); err != nil {
		t.Fatalf("read-only owner ceiling rejected: %v", err)
	}
}

func TestDecodeDocumentRejectsUnknownMember(t *testing.T) {
	catalog := testCatalog(t, FieldEditable)
	encoded, _, err := EncodeDocument(testDocument("text", "auto"), catalog)
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(encoded, []byte(`{"default_locale"`),
		[]byte(`{"authority":"forged","default_locale"`), 1)
	if _, err := DecodeDocument(changed, catalog); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown member error = %v", err)
	}
}

func testCatalog(t *testing.T, mode FieldMode) FieldCatalog {
	t.Helper()
	field := mustField(t, testTextFieldID, fieldregistry.TypeText,
		fieldregistry.ValueTextNFCV1, mode)
	catalog, err := NewFieldCatalog([]Field{field})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func testDocument(widget, interaction string) Document {
	return Document{
		RegistryContract: RegistryContractVersion,
		DefaultLocale:    "en", Locales: []string{"en"},
		Surfaces: []Surface{{
			ID: "018f47f4-7b1b-7cc0-8000-000000000020", Kind: "record_update",
			Translations: []Translation{{Locale: "en", Label: "Update"}},
			Root: Node{
				ID: "018f47f4-7b1b-7cc0-8000-000000000021", Kind: "container",
				Variant: "stack", Span: 1, Translations: []Translation{},
				Children: []Node{{
					ID: "018f47f4-7b1b-7cc0-8000-000000000022", Kind: "field",
					FieldID: testTextFieldID, Widget: widget, Interaction: interaction, Span: 1,
					Translations: []Translation{{Locale: "en", Label: "Email"}},
				}},
			},
		}},
	}
}
