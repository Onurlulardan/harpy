package invoice

import (
	"context"
	"fmt"
	"strings"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	invoicedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/invoice"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/money"
	"github.com/google/uuid"
)

func (storage *store) lockDraft(
	ctx context.Context,
	invoiceID uuid.UUID,
) (invoiceapp.DraftSnapshot, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return invoiceapp.DraftSnapshot{}, err
	}
	result := invoiceapp.DraftSnapshot{}
	var currency string
	var registryVersion int64
	var settledMinor int64
	var policySHA256 []byte
	var linesDocument []byte
	err = storage.queryRow(ctx, lockDraftSQL, authorityID, storage.tenantID, invoiceID).Scan(
		&result.InvoiceID, &result.Revision, &currency, &registryVersion, &settledMinor,
		&policySHA256, &result.Finalized, &linesDocument,
	)
	if err != nil {
		return invoiceapp.DraftSnapshot{}, fmt.Errorf("lock billing invoice draft: %w", err)
	}
	currency = strings.TrimSpace(currency)
	result.Settled, err = money.New(settledMinor, currency, registryVersion)
	if err != nil || len(policySHA256) != len(result.PolicySHA256) {
		return invoiceapp.DraftSnapshot{}, invoiceapp.ErrSnapshot
	}
	copy(result.PolicySHA256[:], policySHA256)
	result.Lines, err = decodeVerifiedLines(linesDocument, currency, registryVersion)
	if err != nil {
		return invoiceapp.DraftSnapshot{}, invoiceapp.ErrSnapshot
	}
	return result, nil
}

func verifiedLines(
	identifiers []uuid.UUID,
	gross []int64,
	discount []int64,
	tax []int64,
	debit []int64,
	credit []int64,
	total []int64,
	currency string,
	registryVersion int64,
) ([]invoicedomain.Line, error) {
	length := len(identifiers)
	for _, values := range [][]int64{gross, discount, tax, debit, credit, total} {
		if len(values) != length {
			return nil, invoicedomain.ErrLine
		}
	}
	lines := make([]invoicedomain.Line, 0, length)
	for index, rawID := range identifiers {
		var identifier invoicedomain.LineIdentifier
		copy(identifier[:], rawID[:])
		values := make([]money.Money, 6)
		for position, units := range []int64{
			gross[index], discount[index], tax[index], debit[index], credit[index], total[index],
		} {
			value, err := money.New(units, currency, registryVersion)
			if err != nil {
				return nil, err
			}
			values[position] = value
		}
		line, err := invoicedomain.NewVerifiedLine(
			identifier, values[0], values[1], values[2], values[3], values[4], values[5],
		)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}
