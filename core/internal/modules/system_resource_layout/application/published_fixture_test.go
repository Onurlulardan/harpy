package application

import (
	"testing"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func fixturePublishedPresentation(
	t *testing.T,
	revision layoutapi.Revision,
	catalog presentation.FieldCatalog,
) PublishedPresentation {
	t.Helper()
	compiled, err := compilePresentation(revision, catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalogJSON, catalogSHA, err := catalogEvidence(catalog)
	if err != nil {
		t.Fatal(err)
	}
	return PublishedPresentation{
		Revision: revision, Generation: 1,
		SourceDefinitionSHA256: revision.DefinitionSHA256,
		CompilerVersion:        compiled.compilerVersion,
		CatalogContract:        presentation.FieldCatalogContract,
		CatalogJSON:            catalogJSON, CatalogSHA256: catalogSHA,
		PresentationContract: presentation.ContractVersion,
		PresentationJSON:     compiled.canonicalJSON,
		PresentationSHA256:   compiled.digest, EvidenceComplete: true,
	}
}
