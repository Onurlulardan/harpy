package application

import (
	"crypto/sha256"
	"testing"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const legacyDefinitionDocument = `{"contract":"harpy/system-resource-layout/v1",` +
	`"sections":[{"fields":[{"field_id":"33333333-3333-4333-8333-333333333333",` +
	`"order":0,"translations":[],"widget":"auto"}],` +
	`"id":"22222222-2222-4222-8222-222222222222","order":0,` +
	`"translations":[{"help":"","label":"General","locale":"en"}]}]}`

func TestDecodeDefinitionPreservesLegacyCompilerContract(t *testing.T) {
	definition, canonical, err := decodeDefinition([]byte(legacyDefinitionDocument))
	if err != nil {
		t.Fatal(err)
	}
	if definition.SourceContract != layoutapi.DefinitionContractV1 ||
		len(definition.Surfaces) != 4 ||
		definition.Surfaces[0].Sections[0].Fields[0].Interaction !=
			layoutapi.InteractionAuto {
		t.Fatalf("legacy definition = %+v", definition)
	}
	revision := fixtureRevision(t, layoutapi.RevisionDraft, 1)
	revision.Definition = definition
	revision.DefinitionSHA256 = sha256.Sum256(canonical)
	compiled, err := compilePresentation(
		revision, fixtureCatalogForIDs(t, fixtureFieldOne),
	)
	if err != nil || compiled.compilerVersion != presentationCompilerVersionV1 ||
		len(compiled.document.Surfaces) != 4 {
		t.Fatalf("legacy compile = %+v: %v", compiled, err)
	}
}

func TestEncodeDefinitionAlwaysEmitsExactV2Contract(t *testing.T) {
	definition, canonical, _, err := encodeDefinition(fixtureDefinition())
	if err != nil {
		t.Fatal(err)
	}
	decoded, decodedCanonical, err := decodeDefinition(canonical)
	if err != nil || definition.SourceContract != layoutapi.DefinitionContractV2 ||
		decoded.SourceContract != layoutapi.DefinitionContractV2 ||
		string(canonical) != string(decodedCanonical) {
		t.Fatalf("v2 round trip = %+v/%+v: %v", definition, decoded, err)
	}
	invalid := append(canonical[:len(canonical)-1], []byte(`,"unknown":true}`)...)
	if _, _, err := decodeDefinition(invalid); err != layoutapi.ErrInvalid {
		t.Fatalf("unknown root field error = %v", err)
	}
}

func TestV2SurfacesCompileIndependentFieldSetsAndOptionalPrint(t *testing.T) {
	definition := fixtureDefinition()
	definition.Surfaces[3].Sections[0].Fields =
		definition.Surfaces[3].Sections[0].Fields[:1]
	definition.Surfaces[4].Sections = []layoutapi.Section{}
	definition.Surfaces = append(definition.Surfaces, layoutapi.SurfaceDefinition{
		Kind: layoutapi.SurfacePrint, Sections: []layoutapi.Section{},
		Translations: []layoutapi.Translation{{Locale: "en", Label: "Print"}},
	})
	normalized, _, digest, err := encodeDefinition(definition)
	if err != nil {
		t.Fatal(err)
	}
	revision := fixtureRevision(t, layoutapi.RevisionDraft, 1)
	revision.Definition, revision.DefinitionSHA256 = normalized, digest
	compiled, err := compilePresentation(
		revision, fixturePresentationCatalog(t, presentation.FieldEditable),
	)
	if err != nil {
		t.Fatal(err)
	}
	surfaces := compiled.document.Surfaces
	if len(surfaces) != 6 ||
		len(surfaces[0].Root.Children[0].Children) != 2 ||
		len(surfaces[3].Root.Children[0].Children) != 1 ||
		len(surfaces[4].Root.Children) != 0 || surfaces[5].Kind != "print" {
		t.Fatalf("independent surfaces = %+v", surfaces)
	}
}
