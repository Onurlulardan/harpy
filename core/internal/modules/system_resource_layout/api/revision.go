package api

import (
	"time"
)

type RevisionState string

const (
	RevisionDraft     RevisionState = "draft"
	RevisionPublished RevisionState = "published"
)

func (state RevisionState) Valid() bool {
	return state == RevisionDraft || state == RevisionPublished
}

type Revision struct {
	ID               string
	Resource         ResourceKind
	Version          int64
	State            RevisionState
	Definition       Definition
	DefinitionSHA256 [32]byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PublishedAt      *time.Time
	RowRevision      int64
}
