package menu

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) UpdateOverride(
	ctx context.Context, scope platformtx.TenantScope, request UpdateRequest,
) (OverrideReceipt, error) {
	if request.ItemID == uuid.Nil ||
		!idempotencyPattern.MatchString(request.IdempotencyKey) ||
		request.Label == nil && request.Order == nil && request.Visible == nil {
		return OverrideReceipt{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationUpdate, permissionManage)
	if err != nil {
		return OverrideReceipt{}, err
	}
	itemID, err := packagecontract.ParseMenuItemID(request.ItemID.String())
	if err != nil {
		return OverrideReceipt{}, ErrInvalid
	}
	override := menudomain.Override{ItemID: itemID, Label: request.Label,
		Order: request.Order, Visible: request.Visible}
	if _, err = menudomain.ApplyOverrides([]menudomain.Item{{
		MenuItem: packagecontract.MenuItem{ID: itemID, Label: "validation"}, Visible: true,
	}}, []menudomain.Override{override}); err != nil {
		return OverrideReceipt{}, ErrInvalid
	}
	digest, err := overrideDigest(request)
	if err != nil {
		return OverrideReceipt{}, err
	}
	now := application.clock.Now()
	receiptID, idErr := application.ids.New()
	if !canonicalTime(now) || idErr != nil || receiptID == uuid.Nil || receiptID.Version() != 7 {
		return OverrideReceipt{}, ErrUnavailable
	}
	return store.Update(ctx, OverrideMutation{Override: override,
		ReceiptID: receiptID, ExpectedRevision: request.ExpectedRevision, ActorID: identity.ActorID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest, OccurredAt: now})
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func overrideDigest(value UpdateRequest) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, 4096)
	if err != nil {
		return "", ErrInvalid
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
