package postgres

import (
	"context"
	"crypto/sha256"
	"testing"

	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func TestPublishDecodesCompiledPresentationEvidence(t *testing.T) {
	store := fixtureLayoutStore(fixtureRow{
		document: fixturePublicationJSON(t, true),
	})
	catalogJSON := []byte(`{"contract":"forms-presentation-field-catalog.v1","fields":[]}`)
	presentationJSON := []byte(`{"registry_contract":"forms-presentation-registry.v1"}`)
	published, err := store.Publish(context.Background(), layoutapp.PublishStoreInput{
		ID: adapterRevisionID, ExpectedRowRevision: 1,
		SourceDefinitionSHA256: sha256.Sum256([]byte(adapterDefinition)),
		CompilerVersion:        "system-resource-presentation-compiler.v2",
		CatalogContract:        "forms-presentation-field-catalog.v1",
		CatalogJSON:            catalogJSON, CatalogSHA256: sha256.Sum256(catalogJSON),
		PresentationContract: "forms-presentation.v1",
		PresentationJSON:     presentationJSON,
		PresentationSHA256:   sha256.Sum256(presentationJSON), Now: adapterNow,
	})
	if err != nil || !published.EvidenceComplete ||
		published.Revision.PublishedAt == nil {
		t.Fatalf("published evidence = %+v: %v", published, err)
	}
}
