package application

import (
	"bytes"
	"crypto/sha256"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

type StoredRevisionInput struct {
	ID               uuid.UUID
	Resource         layoutapi.ResourceKind
	Version          int64
	State            layoutapi.RevisionState
	DefinitionJSON   []byte
	DefinitionSHA256 []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PublishedAt      *time.Time
	RowRevision      int64
}

func HydrateRevision(input StoredRevisionInput) (layoutapi.Revision, error) {
	if !stableUUID(input.ID) || !input.Resource.Valid() || input.Version < 1 ||
		!input.State.Valid() || input.RowRevision < 1 || input.CreatedAt.IsZero() ||
		input.UpdatedAt.Before(input.CreatedAt) || len(input.DefinitionSHA256) != sha256.Size {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	definition, canonical, err := decodeDefinition(input.DefinitionJSON)
	if err != nil {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	digest := sha256.Sum256(canonical)
	if !bytes.Equal(digest[:], input.DefinitionSHA256) ||
		(input.State == layoutapi.RevisionDraft && input.PublishedAt != nil) ||
		(input.State == layoutapi.RevisionPublished &&
			(input.PublishedAt == nil || input.PublishedAt.Before(input.CreatedAt))) {
		return layoutapi.Revision{}, layoutapi.ErrInvalid
	}
	publishedAt := cloneTime(input.PublishedAt)
	return layoutapi.Revision{
		ID: input.ID.String(), Resource: input.Resource, Version: input.Version,
		State: input.State, Definition: definition, DefinitionSHA256: digest,
		CreatedAt: input.CreatedAt, UpdatedAt: input.UpdatedAt,
		PublishedAt: publishedAt, RowRevision: input.RowRevision,
	}, nil
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
