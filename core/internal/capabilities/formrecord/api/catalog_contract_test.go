package recordkernel

import (
	"bytes"
	"github.com/google/uuid"
	"testing"
)

func TestV1AndV2CompileThroughOneEnvelopeAuthority(t *testing.T) {
	v1, err := CompileCatalog(fixtureCompileInput(t, DefinitionV1Contract))
	if err != nil {
		t.Fatal(err)
	}
	if v1.AllowsWorkflow(ActionCreate) {
		t.Fatal("V1 invented workflow authority")
	}
	v2Input := fixtureCompileInput(t, DefinitionV2Contract)
	if v2Input.Projection.OriginalContract() != DefinitionV2Contract {
		t.Fatal("V2 identity was lost")
	}
	v2, err := CompileCatalog(v2Input)
	if err != nil {
		t.Fatal(err)
	}
	if !v2.AllowsWorkflow(ActionCreate) || v1.SHA256() == v2.SHA256() {
		t.Fatal("V2 record contract was not bound")
	}
}

func TestCatalogCanonicalRoundTripPreservesRuntimeSemantics(t *testing.T) {
	input := fixtureCompileInput(t, DefinitionV2Contract)
	compiled, err := CompileCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCatalog(DecodeCatalogInput{
		CanonicalJSON: compiled.CanonicalJSON(), ExpectedSHA256: compiled.SHA256(),
		Binding: input.Binding,
	})
	if err != nil {
		t.Fatal(err)
	}
	if decoded.SHA256() != compiled.SHA256() ||
		!bytes.Equal(decoded.CanonicalJSON(), compiled.CanonicalJSON()) {
		t.Fatal("catalog bytes changed during decode")
	}
	authority := externalActorAuthority(t)
	leftInput, err := DecodeInput(compiled, ActionCreate,
		authority, []byte(`{"title":"ok"}`))
	if err != nil {
		t.Fatal(err)
	}
	rightInput, err := DecodeInput(decoded, ActionCreate,
		authority, []byte(`{"title":"ok"}`))
	if err != nil {
		t.Fatal(err)
	}
	left, leftErr := ExecuteMutation(compiled, Record{}, leftInput)
	right, rightErr := ExecuteMutation(decoded, Record{}, rightInput)
	if leftErr != nil || rightErr != nil ||
		!bytes.Equal(left.CanonicalJSON(), right.CanonicalJSON()) ||
		left.Instructions() != right.Instructions() {
		t.Fatalf("semantic mismatch: %v / %v", leftErr, rightErr)
	}
	tampered := compiled.CanonicalJSON()
	tampered[len(tampered)-2] ^= 1
	if _, err := DecodeCatalog(DecodeCatalogInput{
		CanonicalJSON: tampered, ExpectedSHA256: compiled.SHA256(),
		Binding: input.Binding,
	}); err == nil {
		t.Fatal("tampered catalog was accepted")
	}
}

func externalActorAuthority(t *testing.T) MutationAuthority {
	t.Helper()
	return NewActorAuthority()
}

func TestCatalogRejectsHashTransplantsAndBindsPublicationIdentity(t *testing.T) {
	input := fixtureCompileInput(t, DefinitionV2Contract)
	base, err := CompileCatalog(input)
	if err != nil {
		t.Fatal(err)
	}
	tamperedProjection := input
	tamperedProjection.Binding.SourceDocumentSHA256[0] ^= 0xff
	if _, err := CompileCatalog(tamperedProjection); err == nil {
		t.Fatal("source digest transplant was accepted")
	}
	tamperedField := input
	tamperedField.Binding.Fields = append([]PublishedField(nil), input.Binding.Fields...)
	tamperedField.Binding.Fields[0].PublicName = "forged"
	if _, err := CompileCatalog(tamperedField); err == nil {
		t.Fatal("field catalog transplant was accepted")
	}
	tamperedRegistry := input
	tamperedRegistry.Registry.Snapshot.Types = append(
		[]Descriptor(nil), input.Registry.Snapshot.Types...,
	)
	tamperedRegistry.Registry.Snapshot.Types[0].ValueProfile = "forged_v1"
	if _, err := CompileCatalog(tamperedRegistry); err == nil {
		t.Fatal("registry transplant was accepted")
	}
	rebound := input
	rebound.Binding.PublicationID = uuid.MustParse(
		"018f47f4-7b1b-7cc0-8000-000000001099",
	)
	rebound.Binding.PublicationRevision = 2
	other, err := CompileCatalog(rebound)
	if err != nil {
		t.Fatal(err)
	}
	if base.SHA256() == other.SHA256() {
		t.Fatal("publication identity was absent from catalog attestation")
	}
}

func TestCatalogAcceptsAdoptedV4TenantButRejectsInvalidTenantIdentity(t *testing.T) {
	input := fixtureCompileInput(t, DefinitionV1Contract)
	if input.Binding.TenantID.Version() != 4 {
		t.Fatal("fixture is not an adopted UUIDv4 tenant")
	}
	catalog, err := CompileCatalog(input)
	if err != nil || catalog.TenantID() != input.Binding.TenantID {
		t.Fatalf("V4 tenant compile = %s, %v", catalog.TenantID(), err)
	}
	if _, err := DecodeCatalog(DecodeCatalogInput{
		CanonicalJSON: catalog.CanonicalJSON(), ExpectedSHA256: catalog.SHA256(),
		Binding: input.Binding,
	}); err != nil {
		t.Fatalf("V4 tenant decode: %v", err)
	}
	input.Binding.TenantID = uuid.Nil
	if _, err := CompileCatalog(input); err == nil {
		t.Fatal("nil tenant was accepted")
	}
}
