package iam

import (
	"context"
	"encoding/hex"
	"errors"

	permissioncacheapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissioncache"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	ErrPermissionCacheInvalid     = permissioncacheapp.ErrInvalid
	ErrPermissionCacheConflict    = permissioncacheapp.ErrConflict
	ErrPermissionCacheUnavailable = permissioncacheapp.ErrUnavailable
)

type PermissionCacheFlushCommand struct {
	TenantID, ActorID, OperationID                    string
	RequestSHA256                                     string
	TokenVersion, AuthorizationVersion, PlatformClock int64
}

type PermissionCacheFlushResult struct {
	AuthorizationClock, Revision int64
	Replayed                     bool
}

func (runtime *PermissionCacheRuntime) Flush(
	ctx context.Context,
	command PermissionCacheFlushCommand,
) (PermissionCacheFlushResult, error) {
	if runtime == nil || runtime.pool == nil || runtime.coordinator == nil ||
		runtime.application == nil || ctx == nil || ctx.Err() != nil {
		return PermissionCacheFlushResult{}, ErrPermissionCacheInvalid
	}
	tenantID, tenantErr := uuid.Parse(command.TenantID)
	actorID, actorErr := uuid.Parse(command.ActorID)
	operationID, operationErr := uuid.Parse(command.OperationID)
	digest, digestErr := hex.DecodeString(command.RequestSHA256)
	if tenantErr != nil || actorErr != nil || operationErr != nil ||
		digestErr != nil || len(digest) != 32 {
		return PermissionCacheFlushResult{}, ErrPermissionCacheInvalid
	}
	var requestSHA [32]byte
	copy(requestSHA[:], digest)
	clear(digest)
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"})
	if err != nil || len(generations) != 2 || generations["iam"] < 1 ||
		generations["iam"] != generations["system_resource_layout"] {
		return PermissionCacheFlushResult{}, ErrPermissionCacheUnavailable
	}
	scopeRequest, err := platformtx.AdmitCommandScope(
		actorID, tenantID, "iam.permission_cache.flush",
		identityAuthorityGroup, generations["iam"], generations)
	if err != nil {
		return PermissionCacheFlushResult{}, ErrPermissionCacheUnavailable
	}
	var result permissioncacheapp.Result
	err = runtime.coordinator.InTenant(ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.application.Purge(scoped, permissioncacheapp.Command{
				Scope: scope, TenantID: tenantID, ActorID: actorID,
				OperationID: operationID, RequestSHA256: requestSHA,
				TokenVersion:         command.TokenVersion,
				AuthorizationVersion: command.AuthorizationVersion,
				PlatformClock:        command.PlatformClock})
			return err
		})
	if err != nil {
		if errors.Is(err, permissioncacheapp.ErrConflict) ||
			errors.Is(err, platformtx.ErrCommitRejected) {
			return PermissionCacheFlushResult{}, ErrPermissionCacheConflict
		}
		return PermissionCacheFlushResult{}, ErrPermissionCacheUnavailable
	}
	return PermissionCacheFlushResult{AuthorizationClock: result.AuthorizationClock,
		Revision: result.Revision, Replayed: result.Replayed}, nil
}
