package recoveryanchor

import (
	"context"
	"encoding/hex"
	"errors"

	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const recoveryAnchorAuthorityGroup = "iam_system_resources"

func parseRuntimeSubject(value anchorapi.Subject) (anchordomain.Subject, error) {
	environmentID, environmentErr := uuid.Parse(value.EnvironmentID)
	tenantID, tenantErr := uuid.Parse(value.TenantID)
	actorID, actorErr := uuid.Parse(value.ActorID)
	if environmentErr != nil || tenantErr != nil || actorErr != nil ||
		environmentID == uuid.Nil || tenantID == uuid.Nil || actorID == uuid.Nil {
		return anchordomain.Subject{}, anchorapi.ErrInvalid
	}
	return anchordomain.Subject{EnvironmentID: environmentID,
		TenantID: tenantID, ActorID: actorID}, nil
}

func parseRuntimeDigest(value string) ([32]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != value {
		clear(decoded)
		return [32]byte{}, anchorapi.ErrInvalid
	}
	var result [32]byte
	copy(result[:], decoded)
	clear(decoded)
	return result, nil
}

func (runtime *Runtime) scopeRequest(
	ctx context.Context, subject anchordomain.Subject, operation string, command bool,
) (platformtx.ScopeRequest, error) {
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, subject.TenantID, []string{"iam", "system_resource_layout"})
	if err != nil || len(generations) != 2 || generations["iam"] < 1 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, anchorapi.ErrUnavailable
	}
	if command {
		request, scopeErr := platformtx.AdmitCommandScope(subject.ActorID,
			subject.TenantID, operation, recoveryAnchorAuthorityGroup,
			generations["iam"], generations)
		return request, mapRuntimeError(scopeErr)
	}
	request, scopeErr := queryruntime.IAMSnapshotScopeRequest(subject.ActorID,
		subject.TenantID, map[string]int64{"iam": generations["iam"]})
	return request, mapRuntimeError(scopeErr)
}

func mapRuntimeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, anchordomain.ErrInvalid):
		return anchorapi.ErrInvalid
	case errors.Is(err, anchordomain.ErrDenied):
		return anchorapi.ErrDenied
	case errors.Is(err, anchordomain.ErrNotFound):
		return anchorapi.ErrNotFound
	case errors.Is(err, anchordomain.ErrConflict),
		errors.Is(err, platformtx.ErrCommitRejected):
		return anchorapi.ErrConflict
	default:
		return anchorapi.ErrUnavailable
	}
}
