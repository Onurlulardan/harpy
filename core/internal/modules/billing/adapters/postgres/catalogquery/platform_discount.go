package catalogquery

import (
	"context"
	"fmt"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	"github.com/google/uuid"
)

func (storage *store) platformDiscount(
	ctx context.Context, id uuid.UUID,
) (catalogapp.DiscountRecord, bool, error) {
	if !storage.usable(ctx) {
		return catalogapp.DiscountRecord{}, false, catalogapp.ErrConfiguration
	}
	resultRows, err := storage.query(ctx, getPlatformDiscountSQL, id)
	if err != nil {
		return catalogapp.DiscountRecord{}, false,
			fmt.Errorf("query platform billing discount: %w", err)
	}
	defer resultRows.Close()
	if !resultRows.Next() {
		if err = resultRows.Err(); err != nil {
			return catalogapp.DiscountRecord{}, false, err
		}
		return catalogapp.DiscountRecord{}, false, nil
	}
	var document []byte
	if err = resultRows.Scan(&document); err != nil {
		return catalogapp.DiscountRecord{}, false,
			fmt.Errorf("scan platform billing discount: %w", err)
	}
	result, err := decodeDiscount(document)
	if err != nil {
		return catalogapp.DiscountRecord{}, false, err
	}
	if resultRows.Next() || resultRows.Err() != nil {
		return catalogapp.DiscountRecord{}, false,
			fmt.Errorf("query platform billing discount: invalid cardinality")
	}
	return result, true, nil
}
