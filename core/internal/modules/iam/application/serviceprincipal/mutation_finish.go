package serviceprincipal

import (
	"context"
	"reflect"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) finish(
	ctx context.Context,
	scope transaction.TenantScope,
	evidence MutationEvidence,
	result MutationResult,
	commandKey string,
	resultCode string,
) (MutationResult, error) {
	if result.Replayed {
		return result, nil
	}
	change := rbacdomain.Change{
		OperationID:   rbacdomain.ID(evidence.OperationID),
		ActorID:       rbacdomain.ID(evidence.ActorID),
		AggregateID:   rbacdomain.ID(result.ID),
		RequestSHA256: evidence.RequestSHA256,
		CommandKey:    commandKey, ResultCode: resultCode,
		Revision: result.Revision, Clock: result.AuthorizationClock,
	}
	if change.Validate() != nil || commands.sink.WriteChange(ctx, scope, change) != nil {
		return MutationResult{}, ErrChangeSink
	}
	return result, nil
}

func changeSinkNil(sink rbacports.ChangeSink[transaction.TenantScope]) bool {
	if sink == nil {
		return true
	}
	value := reflect.ValueOf(sink)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
