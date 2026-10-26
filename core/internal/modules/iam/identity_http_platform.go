package iam

import (
	"context"

	"github.com/google/uuid"
)

type IdentityHTTPPlatform interface {
	ResolveSnapshotDomainGenerations(
		context.Context, uuid.UUID, []string,
	) (map[string]int64, error)
}
