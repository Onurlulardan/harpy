package tenantsettings

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) TestEmail(ctx context.Context,
	scope platformtx.TenantScope, command TestEmailCommand,
) (TestEmailResult, error) {
	command.To = normalizeEmail(command.To)
	if application == nil || ctx == nil || ctx.Err() != nil ||
		scope.TenantID() == uuidNil || !validEmail(command.To) ||
		command.ExpectedRevision < 1 || !validIdempotencyKey(command.IdempotencyKey) {
		return TestEmailResult{}, ErrInvalid
	}
	registry, err := application.registry.Read(ctx, scope)
	if err != nil {
		return TestEmailResult{}, err
	}
	if registry.Revision != command.ExpectedRevision {
		return TestEmailResult{}, ErrConflict
	}
	result, err := application.mail.EnqueueTest(ctx, scope, TestMailRequest{
		TenantID: scope.TenantID(), TenantName: registry.Name, To: command.To,
		ExpectedRevision: command.ExpectedRevision, IdempotencyKey: command.IdempotencyKey,
		RequestSHA256: testEmailDigest(command), RequestedAt: application.clock.Now().UTC()})
	if err != nil {
		return TestEmailResult{}, err
	}
	if result.OperationID == uuidNil {
		return TestEmailResult{}, ErrUnavailable
	}
	return result, nil
}
