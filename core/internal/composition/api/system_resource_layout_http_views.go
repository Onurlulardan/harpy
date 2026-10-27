package api

import (
	"encoding/hex"
	"time"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
)

type layoutHTTPRevisionView struct {
	RevisionID       string                `json:"revision_id"`
	Resource         layouts.Resource      `json:"resource"`
	Version          int64                 `json:"version"`
	State            layouts.RevisionState `json:"state"`
	Definition       layoutHTTPDefinition  `json:"definition"`
	DefinitionSHA256 string                `json:"definition_sha256"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	PublishedAt      *time.Time            `json:"published_at"`
	RowRevision      int64                 `json:"row_revision"`
}

func layoutHTTPRevisionResponse(
	resource layouts.Resource,
	revision layouts.Revision,
) layoutHTTPRevisionView {
	return layoutHTTPRevisionView{
		RevisionID: revision.ID, Resource: resource, Version: revision.Version,
		State: revision.State, Definition: layoutHTTPDefinition{
			Contract: revision.Definition.SourceContract,
			Surfaces: revision.Definition.Surfaces,
		},
		DefinitionSHA256: hex.EncodeToString(revision.DefinitionSHA256[:]),
		CreatedAt:        revision.CreatedAt, UpdatedAt: revision.UpdatedAt,
		PublishedAt: revision.PublishedAt, RowRevision: revision.RowRevision,
	}
}
