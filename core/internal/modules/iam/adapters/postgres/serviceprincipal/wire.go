package serviceprincipal

import (
	"time"

	"github.com/google/uuid"
)

type evidenceDocument struct {
	Principal principalDocument `json:"principal"`
	Grant     grantDocument     `json:"grant"`
}

type principalDocument struct {
	ID                   uuid.UUID `json:"id"`
	Kind                 string    `json:"kind"`
	State                string    `json:"state"`
	AuthorizationVersion int64     `json:"authorization_version"`
	Revision             int64     `json:"revision"`
}

type grantDocument struct {
	ID               uuid.UUID  `json:"id"`
	Revision         int64      `json:"revision"`
	PreviousRevision int64      `json:"previous_revision"`
	PrincipalID      uuid.UUID  `json:"principal_id"`
	Capability       string     `json:"capability"`
	ScopeKind        string     `json:"scope_kind"`
	Owner            string     `json:"owner"`
	ResourceKind     string     `json:"resource_kind"`
	ResourceID       *uuid.UUID `json:"resource_id"`
	ValidFrom        time.Time  `json:"valid_from"`
	ValidUntil       *time.Time `json:"valid_until"`
	State            string     `json:"state"`
	ManifestSHA256   string     `json:"manifest_sha256"`
}
