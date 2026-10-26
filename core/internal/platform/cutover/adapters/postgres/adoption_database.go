package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

type AdoptionDatabase struct {
	source         []*pgx.Conn
	target         []*pgx.Conn
	cleanupTimeout time.Duration
	applyRegistry  platformcutover.ApplyRegistry
	sealVerified   bool
}

func OpenAdoptionDatabase(
	ctx context.Context,
	sourceSettings config.PrivilegedPostgres,
	targetSettings config.PrivilegedPostgres,
	lease attestation.VerifiedCutoverCredentialLease,
	applyRegistry platformcutover.ApplyRegistry,
) (*AdoptionDatabase, error) {
	if lease.SourceConnectionLimit() < 2 || lease.TargetConnectionLimit() < 1 ||
		!applyRegistry.Valid() {
		return nil, fmt.Errorf("cutover lease connection limits are invalid")
	}
	applicationName := cutoverAdoptionApplicationName(lease.RunID())
	database := &AdoptionDatabase{
		cleanupTimeout: targetSettings.CleanupTimeout(), applyRegistry: applyRegistry,
	}
	var err error
	database.source, err = openConnections(
		ctx, sourceSettings, applicationName, lease.SourceConnectionLimit(),
	)
	if err == nil {
		database.target, err = openConnections(
			ctx, targetSettings, applicationName, lease.TargetConnectionLimit(),
		)
	}
	if err != nil {
		database.closeBounded()
		return nil, err
	}
	if err = verifyAdoptionDatabaseIdentity(ctx, database.source, lease.Source()); err == nil {
		err = verifyAdoptionDatabaseIdentity(ctx, database.target, lease.Target())
	}
	if err != nil {
		database.closeBounded()
		return nil, err
	}
	if err := verifyLeasedRoleValidity(
		ctx, database.source[0], "harpy_cutover_decoder", lease.ExpiresAt(),
		true, lease.SourceConnectionLimit(),
	); err == nil {
		err = verifyLeasedRoleValidity(
			ctx, database.target[0], "harpy_adopter", lease.ExpiresAt(),
			false, lease.TargetConnectionLimit(),
		)
	}
	if err != nil {
		database.closeBounded()
		return nil, err
	}
	return database, nil
}

func openConnections(
	ctx context.Context,
	settings config.PrivilegedPostgres,
	applicationName string,
	count int,
) ([]*pgx.Conn, error) {
	connections := make([]*pgx.Conn, 0, count)
	for len(connections) < count {
		connection, err := openConnection(ctx, settings, applicationName)
		if err != nil {
			_ = closeConnections(ctx, connections)
			return nil, err
		}
		connections = append(connections, connection)
	}
	return connections, nil
}

func (database *AdoptionDatabase) Close() error {
	if database == nil {
		return nil
	}
	return database.closeBounded()
}

func (database *AdoptionDatabase) closeBounded() error {
	ctx, cancel := context.WithTimeout(context.Background(), database.cleanupTimeout)
	defer cancel()
	sourceErr := closeConnections(ctx, database.source)
	targetErr := closeConnections(ctx, database.target)
	database.source, database.target = nil, nil
	if sourceErr != nil || targetErr != nil {
		return fmt.Errorf("close cutover PostgreSQL sessions")
	}
	return nil
}
