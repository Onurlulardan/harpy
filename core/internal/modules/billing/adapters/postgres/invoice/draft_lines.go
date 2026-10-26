package invoice

import (
	"encoding/json"

	invoicedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/invoice"
	"github.com/google/uuid"
)

type draftLineDocument struct {
	ID       uuid.UUID `json:"id"`
	Gross    int64     `json:"gross"`
	Discount int64     `json:"discount"`
	Tax      int64     `json:"tax"`
	Debit    int64     `json:"debit"`
	Credit   int64     `json:"credit"`
	Total    int64     `json:"total"`
}

func decodeVerifiedLines(
	raw []byte,
	currency string,
	registryVersion int64,
) ([]invoicedomain.Line, error) {
	var documents []draftLineDocument
	if len(raw) == 0 || json.Unmarshal(raw, &documents) != nil {
		return nil, invoicedomain.ErrLine
	}
	identifiers := make([]uuid.UUID, len(documents))
	values := make([][]int64, 6)
	for index := range values {
		values[index] = make([]int64, len(documents))
	}
	for index, document := range documents {
		identifiers[index] = document.ID
		for position, amount := range []int64{
			document.Gross, document.Discount, document.Tax,
			document.Debit, document.Credit, document.Total,
		} {
			values[position][index] = amount
		}
	}
	return verifiedLines(
		identifiers, values[0], values[1], values[2], values[3], values[4], values[5],
		currency, registryVersion,
	)
}
