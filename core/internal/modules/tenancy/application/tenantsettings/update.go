package tenantsettings

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Update(ctx context.Context, scope platformtx.TenantScope,
	command UpdateCommand,
) (UpdateResult, error) {
	if command.SMTP != nil {
		defer clear(command.SMTP.Password)
	}
	command, valid := normalizeUpdate(command)
	if application == nil || ctx == nil || ctx.Err() != nil || !valid ||
		scope.TenantID() == uuidNil {
		return UpdateResult{}, ErrInvalid
	}
	return application.applyUpdate(ctx, scope, command)
}

func (application *Application) applyUpdate(ctx context.Context,
	scope platformtx.TenantScope, command UpdateCommand,
) (UpdateResult, error) {
	digest := updateDigest(command)
	result, err := application.registry.Update(ctx, scope, RegistryMutation{
		Name: command.Name, ExpectedRevision: command.ExpectedRevision,
		IdempotencyKey: command.IdempotencyKey, RequestSHA256: digest})
	if err != nil || command.SMTP == nil {
		return result, err
	}
	mutation := MailMutation{SMTP: *command.SMTP,
		ExpectedRevision: command.ExpectedRevision,
		ResultRevision:   result.Revision,
		IdempotencyKey:   command.IdempotencyKey, RequestSHA256: digest}
	if err = application.mail.Replace(ctx, scope, mutation); err != nil {
		return UpdateResult{}, err
	}
	return result, nil
}
