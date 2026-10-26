package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Append(
	ctx context.Context, scope platformtx.TenantScope, intent AppendIntent,
) (domain.Entry, error) {
	if intent.OwnerOperation == "" || intent.OwnerOperation != scope.OperationID() {
		return domain.Entry{}, ErrForbidden
	}
	store, identity, err := application.bind(
		ctx, scope, ScopeCommand, intent.OwnerOperation, nil)
	if err != nil {
		return domain.Entry{}, err
	}
	id, err := application.newID()
	if err != nil {
		return domain.Entry{}, err
	}
	recordedAt, err := application.now()
	if err != nil {
		return domain.Entry{}, err
	}
	entry := domain.Entry{ID: domainID(id), OwnerAuditID: domainID(intent.OwnerAuditID),
		ActorID: domainID(identity.ActorID), OrganizationID: optionalDomainID(intent.OrganizationID),
		CorrelationID: optionalDomainID(intent.CorrelationID), OwnerModule: intent.OwnerModule,
		Operation:  intent.OwnerOperation,
		EntityType: intent.EntityType, EntityID: intent.EntityID, FormSlug: intent.FormSlug,
		Action: intent.Action, AuthorizationVersion: intent.AuthorizationVersion,
		ReasonCode:   intent.ReasonCode,
		FieldChanges: append([]domain.FieldChange(nil), intent.FieldChanges...),
		OccurredAt:   intent.OccurredAt.UTC().Truncate(1000), RecordedAt: recordedAt}
	entry.EventSHA256 = requestDigest(entry)
	if entry.Validate() != nil {
		return domain.Entry{}, ErrInvalid
	}
	persisted, _, err := store.Append(ctx, AppendMutation{
		Entry: entry, RequestSHA256: entry.EventSHA256})
	if err != nil {
		return domain.Entry{}, err
	}
	if persisted.Validate() != nil || persisted.OwnerAuditID != entry.OwnerAuditID ||
		persisted.EventSHA256 != entry.EventSHA256 {
		return domain.Entry{}, ErrUnavailable
	}
	return persisted, nil
}
