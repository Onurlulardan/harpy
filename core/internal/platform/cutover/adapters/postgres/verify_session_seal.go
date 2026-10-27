package postgres

import (
	"context"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

func (database *AdoptionDatabase) VerifySessionSeal(
	ctx context.Context,
	lease attestation.VerifiedCutoverCredentialLease,
	seal attestation.VerifiedCutoverSessionSeal,
) error {
	if database == nil || len(database.source) != lease.SourceConnectionLimit() ||
		len(database.target) != lease.TargetConnectionLimit() {
		return fmt.Errorf("cutover live session set differs")
	}
	if err := verifySessionSet(
		ctx, database.source, seal.SourceSessions(), lease.Source(),
	); err != nil {
		return err
	}
	if err := verifySessionSet(
		ctx, database.target, seal.TargetSessions(), lease.Target(),
	); err != nil {
		return err
	}
	if err := verifySealedRole(
		ctx, database.source[0], "harpy_cutover_decoder", lease.ExpiresAt(),
	); err != nil {
		return err
	}
	if err := verifySealedRole(
		ctx, database.target[0], "harpy_adopter", lease.ExpiresAt(),
	); err != nil {
		return err
	}
	database.sealVerified = true
	return nil
}

func verifySealedRole(
	ctx context.Context, connection *pgx.Conn, principal string, expiresAt time.Time,
) error {
	var exact bool
	if err := connection.QueryRow(
		ctx, observeSealedRoleSQL, principal, expiresAt,
	).Scan(&exact); err != nil || !exact {
		return fmt.Errorf("cutover sealed role state differs")
	}
	return nil
}

func verifyLeasedRoleValidity(
	ctx context.Context, connection *pgx.Conn, principal string,
	expiresAt time.Time, replication bool, connectionLimit int,
) error {
	var exact bool
	if err := connection.QueryRow(
		ctx, observeRoleLeaseSQL, principal, expiresAt, replication, connectionLimit,
	).Scan(&exact); err != nil || !exact {
		return fmt.Errorf("cutover live role lease differs")
	}
	return nil
}

func verifySessionSet(
	ctx context.Context,
	connections []*pgx.Conn,
	expected []cutoverv1.CutoverSession,
	databaseIdentity cutoverv1.DatabaseIdentity,
) error {
	if len(connections) != len(expected) {
		return fmt.Errorf("cutover live session set differs")
	}
	matched := make([]bool, len(expected))
	for _, connection := range connections {
		observation, err := observeSession(ctx, connection)
		if err != nil || observation.validateDatabase(databaseIdentity) != nil {
			return fmt.Errorf("cutover live session evidence differs")
		}
		match := -1
		for index, session := range expected {
			if !matched[index] && observation.matches(session) {
				match = index
				break
			}
		}
		if match < 0 {
			return fmt.Errorf("cutover live session evidence differs")
		}
		matched[match] = true
	}
	return nil
}
