package discounteligibility

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

func (storage *store) listOfferings(
	ctx context.Context,
	authorityID, versionID uuid.UUID,
) ([]discountdomain.Identifier, error) {
	resultRows, err := storage.query(ctx, listOfferingsSQL, authorityID, versionID)
	if err != nil {
		return nil, fmt.Errorf("%w: list discount offerings", discountapp.ErrQueryStorage)
	}
	defer resultRows.Close()
	result := make([]discountdomain.Identifier, 0)
	for resultRows.Next() {
		var identifier uuid.UUID
		if err = resultRows.Scan(&identifier); err != nil || !validV7(identifier) {
			return nil, discountapp.ErrQueryInvariant
		}
		result = append(result, discountdomain.Identifier(identifier))
	}
	if err = resultRows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate discount offerings", discountapp.ErrQueryStorage)
	}
	return result, nil
}
