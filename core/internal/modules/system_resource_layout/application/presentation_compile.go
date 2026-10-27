package application

import (
	"crypto/sha256"
	"time"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const (
	presentationCompilerVersionV1 = "system-resource-presentation-compiler.v1"
	presentationCompilerVersionV2 = "system-resource-presentation-compiler.v2"
)

type compiledPresentation struct {
	document        presentation.Document
	canonicalJSON   []byte
	digest          [sha256.Size]byte
	compilerVersion string
}

func compilePresentation(
	revision layoutapi.Revision,
	catalog presentation.FieldCatalog,
) (compiledPresentation, error) {
	if revision.ID == "" || revision.CreatedAt.IsZero() || catalog.Validate() != nil {
		return compiledPresentation{}, layoutapi.ErrInvalid
	}
	document, err := compilePresentationDefinition(
		revision.ID, revision.CreatedAt, revision.Definition, catalog,
	)
	if err != nil {
		return compiledPresentation{}, err
	}
	canonical, digest, err := presentation.EncodeDocument(document, catalog)
	if err != nil {
		return compiledPresentation{}, layoutapi.ErrInvalid
	}
	version, err := presentationCompilerVersionFor(revision.Definition)
	if err != nil {
		return compiledPresentation{}, err
	}
	return compiledPresentation{
		document: document, canonicalJSON: canonical, digest: digest,
		compilerVersion: version,
	}, nil
}

func presentationCompilerVersionFor(
	definition layoutapi.Definition,
) (string, error) {
	switch definition.SourceContract {
	case layoutapi.DefinitionContractV1:
		return presentationCompilerVersionV1, nil
	case layoutapi.DefinitionContractV2:
		return presentationCompilerVersionV2, nil
	default:
		return "", layoutapi.ErrInvalid
	}
}

func compilePresentationDefinition(
	seed string,
	createdAt time.Time,
	definition layoutapi.Definition,
	catalog presentation.FieldCatalog,
) (presentation.Document, error) {
	definition = normalizeDefinition(definition)
	if validateDefinition(definition) != nil || catalog.Validate() != nil {
		return presentation.Document{}, layoutapi.ErrInvalid
	}
	surfaces, err := compilePresentationSurfaces(seed, createdAt, definition, catalog)
	if err != nil {
		return presentation.Document{}, err
	}
	return presentation.Document{
		RegistryContract: presentation.RegistryContractVersion,
		DefaultLocale:    "en", Locales: presentationLocales(definition),
		Surfaces: surfaces,
	}, nil
}
