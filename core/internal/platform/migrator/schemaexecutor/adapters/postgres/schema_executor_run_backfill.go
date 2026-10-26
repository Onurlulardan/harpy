package postgres

import (
	"context"
	"encoding/hex"
	"fmt"

	pagev1 "github.com/Onurlulardan/harpy/contracts/formsrecordpage/v1"
	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/jackc/pgx/v5"
)

type schemaExecutorBackfillPage struct {
	checkpoint     *string
	count          int64
	canonicalBytes *int64
	pageSHA256     *string
}

func schemaExecutorRunBackfill(
	ctx context.Context,
	transaction pgx.Tx,
	statement ddl.Statement,
	checkpoint *string,
) (schemaExecutorBackfillPage, error) {
	var selectedCheckpoint, updatedCheckpoint *string
	var selectedCount, updatedCount int64
	var selectedBytes, updatedBytes, nextRowBytes int64
	var selectedSHA, updatedSHA string
	row := transaction.QueryRow(ctx, statement.Text, checkpoint)
	var err error
	if statement.RecordPageEvidence {
		err = row.Scan(&selectedCheckpoint, &selectedCount, &selectedBytes, &selectedSHA,
			&updatedCheckpoint, &updatedCount, &updatedBytes, &updatedSHA, &nextRowBytes)
	} else {
		err = row.Scan(&selectedCheckpoint, &selectedCount,
			&updatedCheckpoint, &updatedCount)
	}
	if err != nil {
		return schemaExecutorBackfillPage{}, err
	}
	if selectedCount < 0 || selectedCount > schemav1.MaximumBackfillBatch ||
		(selectedCount == 0) != (selectedCheckpoint == nil) {
		return schemaExecutorBackfillPage{}, fmt.Errorf("backfill batch result is outside contract")
	}
	if selectedCount != updatedCount ||
		!schemaExecutorEqualOptionalString(selectedCheckpoint, updatedCheckpoint) {
		return schemaExecutorBackfillPage{},
			schemaExecutorBlocked("schema_backfill_write_suppressed", nil)
	}
	result := schemaExecutorBackfillPage{checkpoint: selectedCheckpoint, count: selectedCount}
	if !statement.RecordPageEvidence {
		return result, nil
	}
	if nextRowBytes > pagev1.MaximumRowBytes {
		return schemaExecutorBackfillPage{},
			schemaExecutorBlocked("schema_backfill_record_too_large", nil)
	}
	if selectedBytes != updatedBytes || selectedSHA != updatedSHA ||
		selectedBytes < 0 || selectedBytes > pagev1.MaximumCanonicalBytes ||
		(selectedCount == 0) != (selectedBytes == 0) ||
		!schemaExecutorPageDigest(selectedSHA) {
		return schemaExecutorBackfillPage{}, fmt.Errorf("backfill page evidence differs")
	}
	if selectedCount > 0 {
		result.canonicalBytes = &selectedBytes
		result.pageSHA256 = &selectedSHA
	}
	return result, nil
}

func schemaExecutorPageDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
}

func schemaExecutorEqualOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
