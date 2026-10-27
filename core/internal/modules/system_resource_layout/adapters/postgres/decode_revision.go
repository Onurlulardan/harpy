package postgres

import (
	"encoding/hex"
	"encoding/json"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	"github.com/google/uuid"
)

func decodeRevision(document []byte) (layoutapi.Revision, error) {
	var wire revisionWire
	if decodeStrictJSON(document, &wire) != nil {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	id, err := uuid.Parse(wire.ID)
	if err != nil || id.String() != wire.ID {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	digest, err := hex.DecodeString(wire.DefinitionSHA256)
	if err != nil || hex.EncodeToString(digest) != wire.DefinitionSHA256 {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	revision, err := layoutapp.HydrateRevision(layoutapp.StoredRevisionInput{
		ID: id, Resource: layoutapi.ResourceKind(wire.ResourceKind),
		Version: wire.Version, State: layoutapi.RevisionState(wire.State),
		DefinitionJSON: wire.Definition, DefinitionSHA256: digest,
		CreatedAt: wire.CreatedAt, UpdatedAt: wire.UpdatedAt,
		PublishedAt: wire.PublishedAt, RowRevision: wire.RowRevision,
	})
	if err != nil {
		return layoutapi.Revision{}, layoutUnavailable()
	}
	return revision, nil
}

func decodeRevisionList(document []byte) ([]layoutapi.Revision, error) {
	var entries []json.RawMessage
	if decodeStrictJSON(document, &entries) != nil || entries == nil {
		return nil, layoutUnavailable()
	}
	revisions := make([]layoutapi.Revision, len(entries))
	for index, entry := range entries {
		revision, err := decodeRevision(entry)
		if err != nil {
			return nil, err
		}
		revisions[index] = revision
	}
	return revisions, nil
}
