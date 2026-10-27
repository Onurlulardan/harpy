package recordkernel

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPhysicalCatalogV2SelfRelationUsesSymbolicExactRootBinding(t *testing.T) {
	catalog, mappings, mappingSHA := relationPhysicalFixture(t, formID)
	document, digest, err := CompilePhysicalCatalogDocumentV2(
		catalog, mappingSHA, mappings, nil,
	)
	if err != nil || digest == [32]byte{} {
		t.Fatalf("compile self relation: %v", err)
	}
	wire := string(document)
	for _, required := range []string{
		`"target_binding_kind":"self"`,
		`"target_publication_id":"` + publicationID + `"`,
		`"target_form_id":"` + formID + `"`,
		`"target_form_version_id":"` + versionID + `"`,
	} {
		if !strings.Contains(wire, required) {
			t.Fatalf("self binding misses %s: %s", required, wire)
		}
	}
	// Self target digests would recursively include the physical catalog's own
	// digest. Their structural absence is the closed, fixed-point-free ABI.
	for _, forbidden := range []string{"target_record_catalog_sha256",
		"target_query_catalog_sha256", "target_physical_catalog_sha256"} {
		if strings.Contains(wire, forbidden) {
			t.Fatalf("self binding contains recursive digest %q", forbidden)
		}
	}
}

func TestPhysicalCatalogV2ExternalRelationSealsExactCatalogs(t *testing.T) {
	catalog, mappings, mappingSHA := relationPhysicalFixture(t, externalFormID)
	binding := externalRelationBindingFixture()
	document, _, err := CompilePhysicalCatalogDocumentV2(
		catalog, mappingSHA, mappings, []ExternalRelationCatalogBinding{binding},
	)
	if err != nil {
		t.Fatal(err)
	}
	wire := string(document)
	for _, required := range []string{
		`"target_binding_kind":"external"`, binding.PublicationID.String(),
		binding.FormID.String(), binding.FormVersionID.String(),
		hexDigest(binding.RecordCatalogSHA256), hexDigest(binding.QueryCatalogSHA256),
		hexDigest(binding.PhysicalCatalogSHA256),
	} {
		if !strings.Contains(wire, required) {
			t.Fatalf("external binding misses %s: %s", required, wire)
		}
	}
}

func TestPhysicalCatalogV2RejectsMissingAndExtraExternalBindings(t *testing.T) {
	catalog, mappings, mappingSHA := relationPhysicalFixture(t, externalFormID)
	if _, _, err := CompilePhysicalCatalogDocumentV2(
		catalog, mappingSHA, mappings, nil,
	); err == nil {
		t.Fatal("missing external binding was accepted")
	}
	extra := externalRelationBindingFixture()
	extra.FormID = uuid.MustParse("018f47f4-7b1b-7cc0-8000-000000001099")
	if _, _, err := CompilePhysicalCatalogDocumentV2(
		catalog, mappingSHA, mappings,
		[]ExternalRelationCatalogBinding{externalRelationBindingFixture(), extra},
	); err == nil {
		t.Fatal("unused external binding was accepted")
	}
}

func externalRelationBindingFixture() ExternalRelationCatalogBinding {
	return ExternalRelationCatalogBinding{
		PublicationID:         uuid.MustParse(externalPublicationID),
		FormID:                uuid.MustParse(externalFormID),
		FormVersionID:         uuid.MustParse(externalFormVersionID),
		RecordCatalogSHA256:   sha256.Sum256([]byte("target-record")),
		QueryCatalogSHA256:    sha256.Sum256([]byte("target-query")),
		PhysicalCatalogSHA256: sha256.Sum256([]byte("target-physical")),
	}
}
