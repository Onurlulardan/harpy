package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func executeAuthorityAdoption(
	ctx context.Context, config invocation,
) (result commandResult, resultErr error) {
	profile := *config.Adopt
	now := time.Now().UTC()
	bundle, err := loadAuthorityBundle(profile.authorityArtifactPaths, now)
	if err != nil {
		return commandResult{}, err
	}
	signer, err := loadAuthoritySigner(profile.SigningKeyFile, bundle.root)
	if err != nil {
		return commandResult{}, err
	}
	connection, err := openAuthorityDatabase(ctx, profile.RootDSNFile, config.Environment)
	if err != nil {
		return commandResult{}, err
	}
	defer connection.Close(context.Background())
	operation := bundle.operation.Value()
	if err := preflightRoot(ctx, connection, operation.Target,
		operation.ExecutorSessionUser); err != nil {
		return commandResult{}, err
	}
	stored, found, err := loadStoredAdoptionReceipt(ctx, connection, bundle, signer)
	if err != nil {
		return commandResult{}, err
	}
	if found {
		if err := publishAdoptionReceipt(profile, stored); err != nil {
			return commandResult{}, err
		}
		return adoptionResult(stored, true), nil
	}
	policy := operation.QuiescencePolicy.Roles[0]
	quiescence, err := quiesceMigrator(ctx, connection, policy)
	if err != nil {
		if restoreMigratorAfterFailure(connection, policy) != nil {
			return commandResult{}, fmt.Errorf("authority adoption recovery required")
		}
		return commandResult{}, err
	}
	committed := false
	defer func() {
		if !committed && restoreMigratorAfterFailure(
			connection, quiescence.policy,
		) != nil {
			result = commandResult{}
			resultErr = fmt.Errorf("authority adoption recovery required")
		}
	}()
	signed, err := adoptInTransaction(ctx, connection, bundle, signer, quiescence)
	if err != nil {
		if replay, exists, replayErr := loadStoredAdoptionReceipt(
			ctx, connection, bundle, signer,
		); replayErr == nil && exists {
			committed = true
			if publishErr := publishAdoptionReceipt(profile, replay); publishErr != nil {
				return commandResult{}, publishErr
			}
			return adoptionResult(replay, true), nil
		}
		return commandResult{}, err
	}
	committed = true
	if err := publishAdoptionReceipt(profile, signed); err != nil {
		return commandResult{}, err
	}
	return adoptionResult(signed, false), nil
}

func adoptInTransaction(
	ctx context.Context, connection *pgx.Conn, bundle verifiedAuthorityBundle,
	signer attestation.Signer, quiescence quiescenceRun,
) (signedAdoptionReceipt, error) {
	transaction, err := connection.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority adoption transaction failed")
	}
	defer transaction.Rollback(context.Background())
	operation := bundle.operation.Value()
	if err := lockAdoptionCatalog(ctx, transaction, operation); err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := assertAuthorityNamespace(ctx, transaction, len(operation.Objects)); err != nil {
		return signedAdoptionReceipt{}, err
	}
	preStates, err := readObjectStates(ctx, transaction, operation.Objects)
	if err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := assertPreAdoptionStates(preStates, operation.Objects); err != nil {
		return signedAdoptionReceipt{}, err
	}
	dependencies, err := readDependencyStates(ctx, transaction, operation.Dependencies)
	if err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := assertDependencyStates(dependencies, operation.Dependencies); err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := transferAdoptedObjects(ctx, transaction, operation.Objects); err != nil {
		return signedAdoptionReceipt{}, err
	}
	postStates, err := readObjectStates(ctx, transaction, operation.Objects)
	if err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := assertPostAdoptionStates(postStates, operation.Objects); err != nil {
		return signedAdoptionReceipt{}, err
	}
	afterDependencies, err := readDependencyStates(ctx, transaction, operation.Dependencies)
	if err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := assertDependencyStatesUnchanged(dependencies, afterDependencies); err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := assertAuthorityOwner(ctx, transaction, len(operation.Objects)); err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := restoreMigrator(ctx, transaction, quiescence.policy); err != nil {
		return signedAdoptionReceipt{}, err
	}
	var adoptedAt time.Time
	if err := transaction.QueryRow(ctx, "SELECT pg_catalog.clock_timestamp()").Scan(
		&adoptedAt,
	); err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority adoption clock failed")
	}
	signed, err := buildSignedAdoptionReceipt(ctx, transaction, bundle, signer,
		uuid.NewString(), adoptedAt, postStates, afterDependencies, quiescence)
	if err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := insertAdoptionReceipt(ctx, transaction, signed); err != nil {
		return signedAdoptionReceipt{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return signedAdoptionReceipt{}, fmt.Errorf("authority adoption commit failed")
	}
	return signed, nil
}

func adoptionResult(signed signedAdoptionReceipt, replayed bool) commandResult {
	return commandResult{Action: "adopt_authority",
		OperationID: signed.receipt.OperationID, EvidenceSHA256: signed.digest,
		Replayed: replayed}
}
