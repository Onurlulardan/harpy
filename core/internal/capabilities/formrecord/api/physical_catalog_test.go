package recordkernel

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestCompilePhysicalCatalogDocumentSealsArtifact(t *testing.T) {
	catalog, document, digest, _, mappingSHA := physicalCatalogFixture(t)
	var wire physicalCatalogDocument
	if json.Unmarshal(document, &wire) != nil || len(wire.Handles) != 4 ||
		!compiledPhysicalDocumentValid(document, digest, catalog, mappingSHA) {
		t.Fatal("physical catalog artifact differs")
	}
}

func TestDigestMappingsIsOrderedAndRejectsTransplant(t *testing.T) {
	_, _, _, mappings, digest := physicalCatalogFixture(t)
	mappings[0], mappings[3] = mappings[3], mappings[0]
	reordered, err := DigestMappings(mappings)
	if err != nil || reordered != digest {
		t.Fatal("mapping digest depends on caller order")
	}
	mappings[0].LogicalID = uuid.MustParse("018f47f4-7b1b-7cc0-8000-000000001099")
	if _, err := DigestMappings(mappings); err == nil {
		t.Fatal("transplanted physical identifier was accepted")
	}
}

func TestDecodePhysicalCatalogRejectsCatalogTransplant(t *testing.T) {
	_, physical, physicalSHA, _, mappingSHA := physicalCatalogFixture(t)
	input := fixtureCompileInput(t, DefinitionV2Contract)
	input.Binding.MappingCatalogSHA256 = mappingSHA
	input.Binding.PublicationID = uuid.MustParse(
		"018f47f4-7b1b-7cc0-8000-000000001099",
	)
	other, err := CompileCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	if compiledPhysicalDocumentValid(physical, physicalSHA, other, mappingSHA) {
		t.Fatal("physical catalog transplant was accepted")
	}
}
