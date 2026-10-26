package application

import (
	"errors"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestCompilePresentationProducesTypedStableSurfaces(t *testing.T) {
	catalog := fixturePresentationCatalog(t, presentation.FieldEditable)
	compiled, err := compilePresentation(fixtureRevision(t, layoutapi.RevisionDraft, 1), catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled.document.Surfaces) != 5 || compiled.digest == [32]byte{} ||
		len(compiled.canonicalJSON) == 0 {
		t.Fatalf("compiled presentation = %+v", compiled)
	}
	second, err := compilePresentation(fixtureRevision(t, layoutapi.RevisionDraft, 1), catalog)
	if err != nil || second.digest != compiled.digest ||
		string(second.canonicalJSON) != string(compiled.canonicalJSON) {
		t.Fatalf("non-deterministic compile: %v", err)
	}
	if compiled.document.Surfaces[0].Root.Children[0].Children[0].FieldID !=
		fixtureFieldOne.String() {
		t.Fatal("stable IAM field reference was not preserved")
	}
}

func TestCompilePresentationRejectsIncompatibleWidget(t *testing.T) {
	boolean, err := presentation.NewField(
		fixtureFieldOne.String(), fieldregistry.TypeBoolean,
		fieldregistry.ValueBooleanV1, presentation.FieldEditable,
	)
	if err != nil {
		t.Fatal(err)
	}
	second := fixturePresentationField(t, fixtureFieldTwo.String(), presentation.FieldEditable)
	catalog, err := presentation.NewFieldCatalog([]presentation.Field{boolean, second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := compilePresentation(
		fixtureRevision(t, layoutapi.RevisionDraft, 1), catalog,
	); !errors.Is(err, layoutapi.ErrInvalid) {
		t.Fatalf("incompatible widget error = %v", err)
	}
}

func fixturePresentationCatalog(
	t *testing.T, mode presentation.FieldMode,
) presentation.FieldCatalog {
	t.Helper()
	first := fixturePresentationField(t, fixtureFieldOne.String(), mode)
	second := fixturePresentationField(t, fixtureFieldTwo.String(), mode)
	catalog, err := presentation.NewFieldCatalog([]presentation.Field{first, second})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func fixturePresentationField(
	t *testing.T, id string, mode presentation.FieldMode,
) presentation.Field {
	t.Helper()
	field, err := presentation.NewField(
		id, fieldregistry.TypeText, fieldregistry.ValueTextNFCV1, mode,
	)
	if err != nil {
		t.Fatal(err)
	}
	return field
}
