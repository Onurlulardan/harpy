package application

import (
	"crypto/sha256"
	"crypto/subtle"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func verifyPublishedPresentation(
	published PublishedPresentation,
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
) (presentation.Document, error) {
	revision := published.Revision
	expectedCompiler, compilerErr := presentationCompilerVersionFor(revision.Definition)
	if !published.EvidenceComplete || published.Generation < 1 ||
		revision.ID == "" || revision.Resource != resource ||
		revision.State != layoutapi.RevisionPublished || revision.PublishedAt == nil ||
		compilerErr != nil || published.CompilerVersion != expectedCompiler ||
		published.CatalogContract != presentation.FieldCatalogContract ||
		published.PresentationContract != presentation.ContractVersion {
		return presentation.Document{}, layoutapi.ErrUnavailable
	}
	verificationCatalog, err := publishedVerificationCatalog(
		published, resource, catalog,
	)
	if err != nil {
		return presentation.Document{}, layoutapi.ErrUnavailable
	}
	catalogJSON, catalogSHA, err := catalogEvidence(verificationCatalog)
	if err != nil || !sameDigest(published.CatalogSHA256, catalogSHA) ||
		!sameBytes(published.CatalogJSON, catalogJSON) ||
		sha256.Sum256(published.CatalogJSON) != published.CatalogSHA256 ||
		!sameDigest(published.SourceDefinitionSHA256, revision.DefinitionSHA256) ||
		sha256.Sum256(published.PresentationJSON) != published.PresentationSHA256 {
		return presentation.Document{}, layoutapi.ErrUnavailable
	}
	document, err := presentation.DecodeDocument(
		published.PresentationJSON, verificationCatalog,
	)
	if err != nil {
		return presentation.Document{}, layoutapi.ErrUnavailable
	}
	compiled, err := compilePresentation(revision, verificationCatalog)
	if err != nil || compiled.compilerVersion != published.CompilerVersion ||
		!sameDigest(compiled.digest, published.PresentationSHA256) ||
		!sameBytes(compiled.canonicalJSON, published.PresentationJSON) {
		return presentation.Document{}, layoutapi.ErrUnavailable
	}
	return presentation.CloneDocument(document), nil
}

func catalogEvidence(
	catalog presentation.FieldCatalog,
) ([]byte, [32]byte, error) {
	document, err := catalog.CanonicalJSON()
	if err != nil {
		return nil, [32]byte{}, err
	}
	digest, err := catalog.CanonicalSHA256()
	return document, digest, err
}

func sameBytes(left, right []byte) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare(left, right) == 1
}

func sameDigest(left, right [32]byte) bool {
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}
