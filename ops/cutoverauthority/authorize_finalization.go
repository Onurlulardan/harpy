package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

const authorizeFinalizationSQL = `SELECT cutover_authority.authorize_finalization_v1(
$1::uuid,$2::uuid,$3::uuid,$4::text,$5::bytea,$6::text,$7::bytea,
$8::text,$9::bytea,$10::text,$11::bytea,$12::text,$13::bytea,
$14::text,$15::text,$16::bytea)`

func executeFinalizationAuthorization(
	ctx context.Context, config invocation,
) (commandResult, error) {
	profile := *config.AuthorizeFinalization
	now := time.Now().UTC()
	bundle, err := loadFinalizationBundle(profile, now)
	if err != nil {
		return commandResult{}, err
	}
	connection, err := openAuthorityDatabase(
		ctx, profile.ClusterAdminDSNFile, config.Environment,
	)
	if err != nil {
		return commandResult{}, err
	}
	defer connection.Close(context.Background())
	if err := preflightClusterAdmin(ctx, connection, bundle.adoption.Target(), now); err != nil {
		return commandResult{}, err
	}
	transaction, err := connection.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return commandResult{}, fmt.Errorf("finalization authorization transaction failed")
	}
	defer transaction.Rollback(context.Background())
	_, err = transaction.Exec(ctx, authorizeFinalizationSQL,
		bundle.finalizer.OperationID(), bundle.finalizer.RunID(), bundle.finalizer.LeaseID(),
		bundle.finalizer.ManifestSHA256(), bundle.documents.finalizerLease,
		bundle.adoption.ManifestSHA256(), bundle.documents.adoptionLease,
		bundle.seal.ReceiptSHA256(), bundle.documents.seal,
		bundle.cleanup.ReceiptSHA256(), bundle.documents.cleanup,
		bundle.receipt.ReceiptSHA256(), bundle.documents.authority,
		bundle.finalizer.TerminalState(), bundle.terminal.SHA256(), bundle.documents.terminal,
	)
	if err != nil {
		return commandResult{}, fmt.Errorf("protected finalization authorization failed")
	}
	if err := transaction.Commit(ctx); err != nil {
		return commandResult{}, fmt.Errorf("finalization authorization commit failed")
	}
	return commandResult{Action: "authorize_finalization",
		OperationID:    bundle.finalizer.OperationID(),
		EvidenceSHA256: bundle.terminal.SHA256()}, nil
}
