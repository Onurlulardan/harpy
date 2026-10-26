package application

import (
	"testing"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestDefaultPresentationIsDeterministicAndCatalogComplete(t *testing.T) {
	catalog := fixturePresentationCatalog(t, presentation.FieldEditable)
	first, err := defaultPresentation(layoutapi.ResourceUser, catalog)
	if err != nil {
		t.Fatal(err)
	}
	second, err := defaultPresentation(layoutapi.ResourceUser, catalog)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, firstDigest, err := presentation.EncodeDocument(first, catalog)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, secondDigest, err := presentation.EncodeDocument(second, catalog)
	if err != nil || firstDigest != secondDigest || string(firstJSON) != string(secondJSON) {
		t.Fatalf("default presentation drift: %v", err)
	}
	for _, surface := range first.Surfaces {
		fields := surface.Root.Children[0].Children
		if len(fields) != 2 {
			t.Fatalf("%s field count = %d", surface.Kind, len(fields))
		}
	}
}

func TestDefaultPresentationHonorsReadOnlyOwnerCeiling(t *testing.T) {
	catalog := fixturePresentationCatalog(t, presentation.FieldReadOnly)
	document, err := defaultPresentation(layoutapi.ResourceUser, catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, surface := range document.Surfaces {
		for _, field := range surface.Root.Children[0].Children {
			if field.Interaction != "read_only" {
				t.Fatalf("%s interaction = %s", surface.Kind, field.Interaction)
			}
		}
	}
}
