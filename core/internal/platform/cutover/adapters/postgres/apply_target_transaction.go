package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type targetApplyInput struct {
	payload             cutoverv1.ChangeBatchPayload
	relation            cutoverv1.RelationMapping
	payloadJSON         []byte
	batchSHA256         string
	leaseManifestSHA256 string
}

func applyTargetTransaction(
	ctx context.Context, target *pgx.Conn, cleanupTimeout time.Duration,
	registry platformcutover.ApplyRegistry, input targetApplyInput,
) (cutoverv1.ApplyResult, error) {
	tx, err := target.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return cutoverv1.ApplyResult{}, fmt.Errorf("begin cutover apply transaction")
	}
	defer rollbackBounded(tx, cleanupTimeout)
	prepared, err := prepareBatch(
		ctx, tx, input.payload, input.batchSHA256, input.leaseManifestSHA256,
	)
	if err != nil || prepared.verify(input.relation) != nil {
		return cutoverv1.ApplyResult{}, fmt.Errorf("verify locked cutover procedure binding")
	}
	if prepared.alreadyApplied {
		if err := tx.Commit(ctx); err != nil {
			return cutoverv1.ApplyResult{}, fmt.Errorf("commit idempotent cutover apply")
		}
		return prepared.result, nil
	}
	query, registered := registry.Lookup(prepared.contractID)
	if !registered {
		return cutoverv1.ApplyResult{}, fmt.Errorf("cutover procedure contract is not registered")
	}
	if err := bindTargetApplySchema(ctx, tx, input.relation, prepared); err != nil {
		return cutoverv1.ApplyResult{}, err
	}
	result, err := executeCompositeApply(
		ctx, tx, query, input.payload, input.payloadJSON,
		input.batchSHA256, input.leaseManifestSHA256,
	)
	if err != nil {
		return cutoverv1.ApplyResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return cutoverv1.ApplyResult{}, fmt.Errorf("commit cutover apply transaction")
	}
	return result, nil
}

type preparedBatch struct {
	contractID         string
	procedureOID       uint32
	procedureSignature string
	alreadyApplied     bool
	result             cutoverv1.ApplyResult
}

func prepareBatch(
	ctx context.Context, tx pgx.Tx, payload cutoverv1.ChangeBatchPayload,
	batchSHA256 string, leaseManifestSHA256 string,
) (preparedBatch, error) {
	var value preparedBatch
	var oid int64
	var count pgtype.Int8
	var keyDigest, stateDigest pgtype.Text
	err := tx.QueryRow(ctx, prepareBatchSQL,
		payload.OperationID, payload.RunID, payload.ScopeKind, payload.ScopeID,
		payload.RelationOrdinal, payload.CheckpointKey, payload.Position.BatchOrdinal,
		payload.Position.CommitLSN, payload.Position.SourceXID, batchSHA256,
		leaseManifestSHA256,
	).Scan(&value.contractID, &oid, &value.procedureSignature, &value.alreadyApplied,
		&count, &keyDigest, &stateDigest)
	if err != nil || oid < 1 || oid > 4294967295 {
		return preparedBatch{}, fmt.Errorf("prepare cutover change batch")
	}
	value.procedureOID = uint32(oid)
	if value.alreadyApplied {
		if !count.Valid || !keyDigest.Valid || !stateDigest.Valid {
			return preparedBatch{}, fmt.Errorf("idempotent cutover result is incomplete")
		}
		value.result = cutoverv1.ApplyResult{
			AppliedCount: count.Int64, KeyStateSHA256: keyDigest.String,
			StateSHA256: stateDigest.String,
		}
		if cutoverv1.ValidateApplyResult(value.result) != nil {
			return preparedBatch{}, fmt.Errorf("idempotent cutover result is invalid")
		}
	} else if count.Valid || keyDigest.Valid || stateDigest.Valid {
		return preparedBatch{}, fmt.Errorf("new cutover result is not empty")
	}
	return value, nil
}

func (value preparedBatch) verify(relation cutoverv1.RelationMapping) error {
	if value.contractID != relation.ProcedureContractID ||
		value.procedureOID != relation.ApplyProcedureOID ||
		value.procedureSignature != relation.ApplyProcedureSignature {
		return fmt.Errorf("cutover procedure binding differs")
	}
	return nil
}

func executeCompositeApply(
	ctx context.Context, tx pgx.Tx, query string,
	payload cutoverv1.ChangeBatchPayload, payloadJSON []byte,
	batchSHA256 string, leaseManifestSHA256 string,
) (cutoverv1.ApplyResult, error) {
	var result cutoverv1.ApplyResult
	err := tx.QueryRow(ctx, query,
		payload.OperationID, payload.RunID, payload.ScopeKind, payload.ScopeID,
		payload.RelationOrdinal, payload.CheckpointKey, payload.Position.SnapshotName,
		payload.Position.SnapshotLSN, payload.Position.ReceivedLSN,
		payload.Position.CommitLSN, payload.Position.FinalLSN,
		payload.Position.SourceXID, payload.Position.BatchOrdinal,
		payload.CheckpointState, json.RawMessage(payload.LastKey),
		payloadJSON, batchSHA256, int64(len(payloadJSON)), leaseManifestSHA256,
	).Scan(&result.AppliedCount, &result.KeyStateSHA256, &result.StateSHA256)
	if err != nil || cutoverv1.ValidateApplyResult(result) != nil {
		return cutoverv1.ApplyResult{}, fmt.Errorf("execute cutover composite apply")
	}
	return result, nil
}
