package lifecycle

import (
	"context"
	"crypto/sha256"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type AuthorizationReceipt struct {
	actorID, tenantID                                      uuid.UUID
	permission, operation                                  string
	tokenVersion, authorizationVersion, authorizationClock int64
	permissionID                                           uuid.UUID
	evidenceSHA256, stepUpPurposeSHA256                    [sha256.Size]byte
}

type authorizationReceiptContextKey struct{}

func NewAuthorizationReceipt(
	scope platformtx.TenantScope,
	permission string,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permissionID uuid.UUID,
	evidenceSHA256, stepUpPurposeSHA256 [sha256.Size]byte,
) (AuthorizationReceipt, error) {
	value := AuthorizationReceipt{
		actorID: scope.ActorID(), tenantID: scope.TenantID(),
		permission: permission, operation: scope.OperationID(),
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permissionID: permissionID,
		evidenceSHA256: evidenceSHA256, stepUpPurposeSHA256: stepUpPurposeSHA256,
	}
	if !value.valid() {
		return AuthorizationReceipt{}, ErrForbidden
	}
	return value, nil
}

func WithAuthorizationReceipt(
	ctx context.Context,
	receipt AuthorizationReceipt,
) (context.Context, error) {
	if ctx == nil || ctx.Err() != nil || !receipt.valid() {
		return nil, ErrForbidden
	}
	return context.WithValue(ctx, authorizationReceiptContextKey{}, receipt), nil
}

func authorizationReceiptFromContext(
	ctx context.Context,
	scope platformtx.TenantScope,
	permission string,
) (AuthorizationReceipt, bool) {
	if ctx == nil || ctx.Err() != nil {
		return AuthorizationReceipt{}, false
	}
	value, ok := ctx.Value(authorizationReceiptContextKey{}).(AuthorizationReceipt)
	return value, ok && value.valid() &&
		value.actorID == scope.ActorID() && value.tenantID == scope.TenantID() &&
		value.operation == scope.OperationID() && value.permission == permission
}

func (value AuthorizationReceipt) values() (
	int64, int64, int64, uuid.UUID, [sha256.Size]byte, [sha256.Size]byte,
) {
	return value.tokenVersion, value.authorizationVersion,
		value.authorizationClock, value.permissionID,
		value.evidenceSHA256, value.stepUpPurposeSHA256
}

func (value AuthorizationReceipt) valid() bool {
	return value.actorID != uuid.Nil && value.tenantID != uuid.Nil &&
		value.permission == "catalog.purge" &&
		value.operation == operationPurge &&
		value.tokenVersion > 0 && value.authorizationVersion > 0 &&
		value.authorizationClock > 0 && value.permissionID != uuid.Nil &&
		value.evidenceSHA256 != [sha256.Size]byte{} &&
		value.stepUpPurposeSHA256 != [sha256.Size]byte{}
}
