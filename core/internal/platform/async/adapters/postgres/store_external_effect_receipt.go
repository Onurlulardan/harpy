package postgres

import (
	"context"
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/jackc/pgx/v5"
)

func storeExternalEffectReceipt(
	ctx context.Context,
	transaction pgx.Tx,
	permit ExternalEffectPermit,
	receipt effectsv1.Receipt,
	values externalReceiptValues,
) error {
	var exact bool
	err := transaction.QueryRow(
		ctx,
		storeExternalEffectReceiptSQL,
		externalReceiptArguments(permit, receipt, values)...,
	).Scan(&exact)
	if err != nil {
		return fmt.Errorf("store external effect receipt: %w", err)
	}
	if !exact {
		return ErrExternalEffectReceipt
	}
	return nil
}
