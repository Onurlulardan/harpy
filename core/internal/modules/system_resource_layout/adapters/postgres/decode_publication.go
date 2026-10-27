package postgres

import (
	"encoding/hex"

	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func decodePublishedPresentation(
	document []byte,
) (layoutapp.PublishedPresentation, error) {
	var wire publicationWire
	if decodeStrictJSON(document, &wire) != nil || wire.Generation < 1 {
		return layoutapp.PublishedPresentation{}, layoutUnavailable()
	}
	revision, err := decodeRevision(wire.Revision)
	if err != nil {
		return layoutapp.PublishedPresentation{}, err
	}
	values := []string{
		wire.SourceDefinitionSHA256, wire.CompilerVersion,
		wire.CatalogContract, wire.CatalogJSON, wire.CatalogSHA256,
		wire.PresentationContract, wire.PresentationJSON, wire.PresentationSHA256,
	}
	present := 0
	for _, value := range values {
		if value != "" {
			present++
		}
	}
	result := layoutapp.PublishedPresentation{
		Revision: revision, Generation: wire.Generation,
	}
	if present == 0 {
		return result, nil
	}
	if present != len(values) {
		return layoutapp.PublishedPresentation{}, layoutUnavailable()
	}
	source, err := decodePublicationDigest(wire.SourceDefinitionSHA256)
	if err != nil {
		return layoutapp.PublishedPresentation{}, err
	}
	catalog, err := decodePublicationDigest(wire.CatalogSHA256)
	if err != nil {
		return layoutapp.PublishedPresentation{}, err
	}
	presentation, err := decodePublicationDigest(wire.PresentationSHA256)
	if err != nil {
		return layoutapp.PublishedPresentation{}, err
	}
	result.SourceDefinitionSHA256 = source
	result.CompilerVersion = wire.CompilerVersion
	result.CatalogContract = wire.CatalogContract
	result.CatalogJSON = []byte(wire.CatalogJSON)
	result.CatalogSHA256 = catalog
	result.PresentationContract = wire.PresentationContract
	result.PresentationJSON = []byte(wire.PresentationJSON)
	result.PresentationSHA256 = presentation
	result.EvidenceComplete = true
	return result, nil
}

func decodePublicationDigest(value string) ([32]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != value {
		return [32]byte{}, layoutUnavailable()
	}
	var digest [32]byte
	copy(digest[:], decoded)
	return digest, nil
}
