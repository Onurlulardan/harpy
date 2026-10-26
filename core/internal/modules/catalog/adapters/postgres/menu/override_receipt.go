package menu

import (
	"context"
	"time"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type storedReceipt struct {
	requestSHA256 string
	receipt       menuapp.OverrideReceipt
	completedAt   time.Time
}

func (storage *store) findReceipt(
	ctx context.Context, mutation menuapp.OverrideMutation,
) (storedReceipt, bool, error) {
	value, inserted, err := scanStoredReceipt(storage.queryRow(ctx, findOverrideReceiptSQL,
		mutation.ActorID.String(), mutation.IdempotencyKey), false)
	if err != nil {
		if storage.notFound(err) {
			return storedReceipt{}, false, nil
		}
		return storedReceipt{}, false, unavailable()
	}
	return value, !inserted, nil
}

func (storage *store) recordReceipt(
	ctx context.Context, mutation menuapp.OverrideMutation, result menuapp.OverrideReceipt,
) (bool, error) {
	value, inserted, err := scanStoredReceipt(storage.queryRow(ctx, recordOverrideReceiptSQL,
		mutation.ReceiptID.String(), mutation.ActorID.String(), mutation.IdempotencyKey,
		mutation.RequestSHA256, result.Override.ItemID.String(), result.Revision,
		result.Override.Label, result.Override.Order, result.Override.Visible, mutation.OccurredAt), true)
	if err != nil || value.requestSHA256 != mutation.RequestSHA256 ||
		value.receipt.Revision != result.Revision || !overrideEqual(value.receipt.Override, result.Override) ||
		!value.completedAt.Equal(mutation.OccurredAt) {
		return false, unavailable()
	}
	return inserted, nil
}

func scanStoredReceipt(source row, withInserted bool) (storedReceipt, bool, error) {
	var requestSHA, itemRaw string
	var revision int64
	var label *string
	var order *int64
	var visible *bool
	var completedAt time.Time
	inserted := false
	targets := []any{&requestSHA, &itemRaw, &revision, &label, &order, &visible, &completedAt}
	if withInserted {
		targets = append(targets, &inserted)
	}
	if err := source.Scan(targets...); err != nil {
		return storedReceipt{}, false, err
	}
	completedAt = completedAt.UTC()
	itemID, err := packagecontract.ParseMenuItemID(itemRaw)
	if err != nil || revision < 1 || !canonicalTime(completedAt) ||
		!digestPattern.MatchString(requestSHA) {
		return storedReceipt{}, false, unavailable()
	}
	override := decodeNullableOverride(itemID, label, order, visible)
	return storedReceipt{requestSHA256: requestSHA, completedAt: completedAt,
		receipt: menuapp.OverrideReceipt{Override: override, Revision: uint64(revision)}}, inserted, nil
}

func decodeNullableOverride(
	itemID packagecontract.MenuItemID, label *string,
	order *int64, visible *bool,
) menudomain.Override {
	value := menudomain.Override{ItemID: itemID}
	if label != nil {
		value.Label = label
	}
	if order != nil {
		orderValue := int(*order)
		value.Order = &orderValue
	}
	if visible != nil {
		value.Visible = visible
	}
	return value
}
