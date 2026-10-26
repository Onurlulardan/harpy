package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type targetFinalizationInput struct {
	operationID, runID, leaseID                                                string
	manifestSHA256, adoptionSHA256, sealSHA256, cleanupSHA256, authoritySHA256 string
	finalizerDocument, adoptionDocument, sealDocument                          []byte
	cleanupDocument                                                            []byte
	authorityDocument                                                          []byte
	terminalState, terminalEvidenceSHA256                                      string
	terminalEvidenceDocument                                                   []byte
}

func finalizeTargetTransaction(
	ctx context.Context, target *pgx.Conn, cleanupTimeout time.Duration,
	input targetFinalizationInput,
) (FinalizationResult, error) {
	tx, err := target.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FinalizationResult{}, fmt.Errorf("begin cutover finalization transaction")
	}
	defer rollbackBounded(tx, cleanupTimeout)
	var result FinalizationResult
	err = tx.QueryRow(ctx, finalizeOperationSQL,
		input.operationID, input.runID, input.leaseID, input.manifestSHA256,
		input.finalizerDocument, input.adoptionSHA256, input.adoptionDocument,
		input.sealSHA256,
		input.sealDocument, input.cleanupSHA256,
		input.cleanupDocument, input.authoritySHA256, input.authorityDocument,
		input.terminalState, input.terminalEvidenceSHA256,
		input.terminalEvidenceDocument,
	).Scan(&result.TerminalState, &result.TerminalEvidenceSHA256)
	if err != nil || result.TerminalState != input.terminalState ||
		result.TerminalEvidenceSHA256 != input.terminalEvidenceSHA256 {
		return FinalizationResult{}, fmt.Errorf("execute cutover finalization")
	}
	if err := tx.Commit(ctx); err != nil {
		return FinalizationResult{}, fmt.Errorf("commit cutover finalization")
	}
	return result, nil
}
