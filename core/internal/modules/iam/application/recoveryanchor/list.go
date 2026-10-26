package recoveryanchor

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) List(
	ctx context.Context,
	scope platformtx.TenantScope,
	subject anchordomain.Subject,
) ([]anchordomain.Anchor, error) {
	if service == nil || !scopeValid(scope, subject) {
		return nil, anchordomain.ErrInvalid
	}
	return service.store.List(ctx, scope, subject)
}

func (runtime *Runtime) Anchors(
	ctx context.Context,
	input anchorapi.Subject,
) ([]anchorapi.Anchor, error) {
	if runtime == nil || runtime.service == nil || ctx == nil || ctx.Err() != nil {
		return nil, anchorapi.ErrInvalid
	}
	subject, err := parseRuntimeSubject(input)
	if err != nil {
		return nil, err
	}
	request, err := runtime.scopeRequest(ctx, subject, "iam.recovery_anchor.list", false)
	if err != nil {
		return nil, err
	}
	var stored []anchordomain.Anchor
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			stored, err = runtime.service.List(scoped, scope, subject)
			return err
		})
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	anchors := make([]anchorapi.Anchor, len(stored))
	for index := range stored {
		anchors[index] = anchorapi.Anchor{AnchorID: runtimeAnchorUUID(stored[index].ID),
			Generation:   stored[index].Generation,
			CredentialID: append([]byte(nil), stored[index].CredentialID...),
			RPID:         stored[index].RPID, UserVerification: stored[index].UserVerification,
			BackupEligible: stored[index].BackupEligible,
			BackupState:    stored[index].BackupState, State: stored[index].State}
	}
	return anchors, nil
}

func runtimeAnchorUUID(value anchordomain.ID) string {
	return uuid.UUID(value).String()
}
