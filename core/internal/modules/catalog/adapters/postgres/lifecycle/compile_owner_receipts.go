package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_owner_receipts.sql
var listOwnerReceiptsSQL string

func (storage *store) CompileOwnerReceipts(
	ctx context.Context, operationID uuid.UUID,
) ([]catalogdomain.OwnerReceipt, error) {
	if operationID == uuid.Nil {
		return nil, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return nil, unavailable()
	}
	resultRows, err := storage.query(ctx, listOwnerReceiptsSQL, operationID)
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	result := make([]catalogdomain.OwnerReceipt, 0)
	for resultRows.Next() {
		value, scanErr := scanOwnerReceipt(resultRows)
		if scanErr != nil || value.OperationID.String() != operationID.String() {
			return nil, unavailable()
		}
		result = append(result, value)
	}
	if resultRows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}
