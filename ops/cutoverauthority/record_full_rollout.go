package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func executeFullRolloutRecord(
	ctx context.Context, config invocation,
) (commandResult, error) {
	profile := *config.RecordFullRollout
	now := time.Now().UTC()
	completion, err := loadFullRolloutCompletion(profile.CompletionEvidence)
	if err != nil || completion.RecordedAt.After(now.Add(time.Minute)) ||
		corecutover.EvaluateFullRollout(completion.Evidence) != nil {
		return commandResult{}, fmt.Errorf("full rollout completion evidence failed")
	}
	bundle, err := loadHistoricalAuthorityBundle(profile.authorityArtifactPaths, now)
	if err != nil {
		return commandResult{}, err
	}
	operation := bundle.operation.Value()
	environmentID, err := uuid.Parse(operation.EnvironmentID)
	if err != nil || completion.EnvironmentID != environmentID {
		return commandResult{}, fmt.Errorf("full rollout environment binding differs")
	}
	if err := verifyFullRolloutArtifacts(
		profile.EvidenceArtifacts, completion.Evidence,
	); err != nil {
		return commandResult{}, err
	}
	connection, err := openAuthorityDatabase(
		ctx, profile.ClusterAdminDSNFile, config.Environment,
	)
	if err != nil {
		return commandResult{}, err
	}
	defer connection.Close(context.Background())
	if err := preflightClusterAdmin(ctx, connection, operation.Target, now); err != nil {
		return commandResult{}, err
	}
	transaction, err := connection.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return commandResult{}, fmt.Errorf("full rollout transaction failed")
	}
	defer transaction.Rollback(context.Background())
	digest, err := corecutover.RecordFullRollout(ctx, transaction, completion)
	if err != nil {
		return commandResult{}, fmt.Errorf("protected full rollout record failed")
	}
	if err := transaction.Commit(ctx); err != nil {
		return commandResult{}, fmt.Errorf("full rollout commit failed")
	}
	return commandResult{Action: "record_full_rollout",
		OperationID:    completion.ReceiptID.String(),
		EvidenceSHA256: hex.EncodeToString(digest[:])}, nil
}
