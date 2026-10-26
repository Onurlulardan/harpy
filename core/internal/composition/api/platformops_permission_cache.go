package api

import (
	"context"
	"errors"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
)

type platformOpsPermissionCache struct {
	runtime *iammodule.PermissionCacheRuntime
}

func (adapter *platformOpsPermissionCache) FlushPermissionCache(
	ctx context.Context,
	input platformopsmodule.PermissionCacheFlush,
) (platformopsmodule.PermissionCacheResult, error) {
	if adapter == nil || adapter.runtime == nil {
		return platformopsmodule.PermissionCacheResult{}, platformopsmodule.ErrUnavailable
	}
	result, err := adapter.runtime.Flush(ctx, iammodule.PermissionCacheFlushCommand{
		TenantID: input.TenantID, ActorID: input.ActorID, OperationID: input.OperationID,
		RequestSHA256: input.RequestSHA256, TokenVersion: input.TokenVersion,
		AuthorizationVersion: input.AuthorizationVersion,
		PlatformClock:        input.PlatformClock})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrPermissionCacheInvalid):
			return platformopsmodule.PermissionCacheResult{}, platformopsmodule.ErrInvalid
		case errors.Is(err, iammodule.ErrPermissionCacheConflict):
			return platformopsmodule.PermissionCacheResult{}, platformopsmodule.ErrConflict
		default:
			return platformopsmodule.PermissionCacheResult{}, platformopsmodule.ErrUnavailable
		}
	}
	return platformopsmodule.PermissionCacheResult{
		AuthorizationClock: result.AuthorizationClock,
		Revision:           result.Revision, Replayed: result.Replayed}, nil
}
