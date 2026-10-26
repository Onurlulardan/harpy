package api

import "context"

type PermissionCacheFlush struct {
	TenantID, ActorID, OperationID, RequestSHA256     string
	TokenVersion, AuthorizationVersion, PlatformClock int64
}

type PermissionCacheResult struct {
	AuthorizationClock, Revision int64
	Replayed                     bool
}

type PermissionCacheOperations interface {
	FlushPermissionCache(context.Context, PermissionCacheFlush) (PermissionCacheResult, error)
}
