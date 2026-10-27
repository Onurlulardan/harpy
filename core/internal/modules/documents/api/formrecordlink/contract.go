package formrecordlink

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const (
	SourceOwner      = "FORMS"
	SourceKind       = "form_record"
	Usage            = "LIFECYCLE"
	SourceConvention = "FORMS/form_record/definition_id/record_id/LIFECYCLE"
	MaximumItemCount = 1000
)

type ID [16]byte

type Request struct {
	DefinitionID   ID
	RecordID       ID
	DesiredItemIDs []ID
	IdempotencyKey string
}

func (request Request) Valid() bool {
	if !uuidV7(request.DefinitionID) || !uuidV7(request.RecordID) ||
		len(request.DesiredItemIDs) > MaximumItemCount ||
		len(request.IdempotencyKey) < 8 || len(request.IdempotencyKey) > 256 {
		return false
	}
	for _, itemID := range request.DesiredItemIDs {
		if !uuidV7(itemID) {
			return false
		}
	}
	return true
}

func uuidV7(value ID) bool {
	return value != ID{} && value[6]>>4 == 7 && value[8]>>6 == 2
}

type Registry interface {
	Reconcile(
		context.Context, platformtx.TenantScope, Request,
	) error
}
